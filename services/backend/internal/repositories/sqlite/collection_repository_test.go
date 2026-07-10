package sqlite

import (
	"testing"
	"time"

	"video-archiver/internal/domain"
	"video-archiver/internal/testutil"
)

func newTestCollection(id, name string) *domain.Collection {
	now := time.Now()
	return &domain.Collection{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// createTestVideo stores a job together with video metadata so it qualifies
// as a collection member.
func createTestVideo(t *testing.T, jobs *JobRepository, id string) {
	t.Helper()
	if err := jobs.Create(testutil.CreateTestJob(id, "https://youtube.com/watch?v="+id)); err != nil {
		t.Fatalf("Create(%s) error = %v", id, err)
	}
	meta := testutil.CreateTestVideoMetadata()
	meta.ID = id
	if err := jobs.StoreMetadata(id, meta); err != nil {
		t.Fatalf("StoreMetadata(%s) error = %v", id, err)
	}
}

func TestCollectionRepository_CreateGetList(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewCollectionRepository(db)
	if err := repo.Create(newTestCollection("col-1", "Watch Later")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID("col-1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil || got.Name != "Watch Later" || got.VideoCount != 0 {
		t.Fatalf("GetByID() = %+v, want name 'Watch Later' with 0 videos", got)
	}

	missing, err := repo.GetByID("nope")
	if err != nil {
		t.Fatalf("GetByID(missing) error = %v", err)
	}
	if missing != nil {
		t.Fatalf("GetByID(missing) = %+v, want nil", missing)
	}

	list, err := repo.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List() returned %d collections, want 1", len(list))
	}
}

func TestCollectionRepository_AddVideosOrderAndCount(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	jobs := NewJobRepository(db)
	repo := NewCollectionRepository(db)

	createTestVideo(t, jobs, "vid-1")
	createTestVideo(t, jobs, "vid-2")
	if err := repo.Create(newTestCollection("col-1", "Mix")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.AddVideos("col-1", []string{"vid-2", "vid-1"}); err != nil {
		t.Fatalf("AddVideos() error = %v", err)
	}
	// Re-adding must be idempotent.
	if err := repo.AddVideos("col-1", []string{"vid-1"}); err != nil {
		t.Fatalf("AddVideos(re-add) error = %v", err)
	}

	videos, err := repo.GetVideos("col-1")
	if err != nil {
		t.Fatalf("GetVideos() error = %v", err)
	}
	if len(videos) != 2 {
		t.Fatalf("GetVideos() returned %d videos, want 2", len(videos))
	}
	// Insertion order is preserved.
	if videos[0].Job.ID != "vid-2" || videos[1].Job.ID != "vid-1" {
		t.Errorf("GetVideos() order = [%s, %s], want [vid-2, vid-1]",
			videos[0].Job.ID, videos[1].Job.ID)
	}

	got, err := repo.GetByID("col-1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.VideoCount != 2 {
		t.Errorf("VideoCount = %d, want 2", got.VideoCount)
	}
	if got.Thumbnail == "" {
		t.Errorf("Thumbnail not derived from first member")
	}
}

func TestCollectionRepository_AddVideosRejectsNonVideos(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	jobs := NewJobRepository(db)
	repo := NewCollectionRepository(db)

	// A job without video metadata (e.g. a playlist parent) must be rejected.
	if err := jobs.Create(testutil.CreateTestJob("playlist-1", "https://youtube.com/playlist?list=x")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.Create(newTestCollection("col-1", "Mix")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.AddVideos("col-1", []string{"playlist-1"}); err == nil {
		t.Fatalf("AddVideos() accepted a non-video job")
	}
	if err := repo.AddVideos("missing-collection", []string{"playlist-1"}); err == nil {
		t.Fatalf("AddVideos() accepted a missing collection")
	}
}

func TestCollectionRepository_RemoveVideoAndListForVideo(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	jobs := NewJobRepository(db)
	repo := NewCollectionRepository(db)

	createTestVideo(t, jobs, "vid-1")
	if err := repo.Create(newTestCollection("col-1", "Mix")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.AddVideos("col-1", []string{"vid-1"}); err != nil {
		t.Fatalf("AddVideos() error = %v", err)
	}

	ids, err := repo.ListForVideo("vid-1")
	if err != nil {
		t.Fatalf("ListForVideo() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != "col-1" {
		t.Fatalf("ListForVideo() = %v, want [col-1]", ids)
	}

	if err := repo.RemoveVideo("col-1", "vid-1"); err != nil {
		t.Fatalf("RemoveVideo() error = %v", err)
	}
	videos, err := repo.GetVideos("col-1")
	if err != nil {
		t.Fatalf("GetVideos() error = %v", err)
	}
	if len(videos) != 0 {
		t.Errorf("GetVideos() returned %d videos after removal, want 0", len(videos))
	}
}

func TestCollectionRepository_DeleteCollection(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	jobs := NewJobRepository(db)
	repo := NewCollectionRepository(db)

	createTestVideo(t, jobs, "vid-1")
	if err := repo.Create(newTestCollection("col-1", "Mix")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.AddVideos("col-1", []string{"vid-1"}); err != nil {
		t.Fatalf("AddVideos() error = %v", err)
	}

	if err := repo.Delete("col-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	got, err := repo.GetByID("col-1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != nil {
		t.Fatalf("collection still exists after Delete()")
	}
	ids, err := repo.ListForVideo("vid-1")
	if err != nil {
		t.Fatalf("ListForVideo() error = %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("memberships not cleaned up on Delete(): %v", ids)
	}

	// The member video itself must survive.
	if _, err := jobs.GetByID("vid-1"); err != nil {
		t.Errorf("member video was deleted with the collection: %v", err)
	}
}

func TestJobRepository_DeleteJobRemovesCollectionMemberships(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	jobs := NewJobRepository(db)
	repo := NewCollectionRepository(db)

	createTestVideo(t, jobs, "vid-1")
	if err := repo.Create(newTestCollection("col-1", "Mix")); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.AddVideos("col-1", []string{"vid-1"}); err != nil {
		t.Fatalf("AddVideos() error = %v", err)
	}

	if err := jobs.DeleteJob("vid-1"); err != nil {
		t.Fatalf("DeleteJob() error = %v", err)
	}

	videos, err := repo.GetVideos("col-1")
	if err != nil {
		t.Fatalf("GetVideos() error = %v", err)
	}
	if len(videos) != 0 {
		t.Errorf("deleted job still member of collection")
	}
}
