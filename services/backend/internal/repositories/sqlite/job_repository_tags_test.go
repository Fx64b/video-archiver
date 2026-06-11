package sqlite

import (
	"testing"
	"video-archiver/internal/domain"
	"video-archiver/internal/testutil"
)

func tagNames(tags []domain.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	return names
}

func containsName(tags []domain.Tag, name string) bool {
	for _, t := range tags {
		if t.Name == name {
			return true
		}
	}
	return false
}

func TestJobRepository_AddAndGetTags(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("job-1", "https://youtube.com/watch?v=test")
	if err := repo.Create(job); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	tags, err := repo.AddTagsToJob("job-1", []string{"music", "Favorites"}, domain.TagSourceUser)
	if err != nil {
		t.Fatalf("AddTagsToJob() error = %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("AddTagsToJob() returned %d tags, want 2: %v", len(tags), tagNames(tags))
	}

	// Re-adding with different casing must not create duplicates.
	tags, err = repo.AddTagsToJob("job-1", []string{"MUSIC"}, domain.TagSourceUser)
	if err != nil {
		t.Fatalf("AddTagsToJob() error = %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("case-insensitive re-add created duplicate: %v", tagNames(tags))
	}
}

func TestJobRepository_RemoveTagPrunesOrphans(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("job-1", "https://youtube.com/watch?v=test")
	repo.Create(job)

	tags, err := repo.AddTagsToJob("job-1", []string{"solo-tag"}, domain.TagSourceUser)
	if err != nil {
		t.Fatalf("AddTagsToJob() error = %v", err)
	}

	if err := repo.RemoveTagFromJob("job-1", tags[0].ID); err != nil {
		t.Fatalf("RemoveTagFromJob() error = %v", err)
	}

	all, err := repo.ListTags()
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}
	if containsName(all, "solo-tag") {
		t.Errorf("orphaned tag was not pruned: %v", tagNames(all))
	}
}

func TestJobRepository_AutoTagsOnStoreMetadata(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)
	job := testutil.CreateTestJob("job-1", "https://youtube.com/watch?v=test")
	repo.Create(job)

	// Test metadata has Categories=[Entertainment], Channel=Test Channel,
	// Tags=[tag1 tag2].
	if err := repo.StoreMetadata("job-1", testutil.CreateTestVideoMetadata()); err != nil {
		t.Fatalf("StoreMetadata() error = %v", err)
	}

	tags, err := repo.GetTagsForJob("job-1")
	if err != nil {
		t.Fatalf("GetTagsForJob() error = %v", err)
	}
	for _, want := range []string{"Entertainment", "Test Channel", "tag1", "tag2"} {
		if !containsName(tags, want) {
			t.Errorf("auto tags missing %q, got: %v", want, tagNames(tags))
		}
	}
	for _, tag := range tags {
		if tag.Source != domain.TagSourceAuto {
			t.Errorf("tag %q has source %q, want auto", tag.Name, tag.Source)
		}
	}

	// Storing metadata again must not duplicate tags.
	before := len(tags)
	if err := repo.StoreMetadata("job-1", testutil.CreateTestVideoMetadata()); err != nil {
		t.Fatalf("StoreMetadata() second call error = %v", err)
	}
	tags, _ = repo.GetTagsForJob("job-1")
	if len(tags) != before {
		t.Errorf("re-storing metadata changed tag count: %d -> %d", before, len(tags))
	}
}

func TestJobRepository_SearchAndTagFilter(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	makeVideo := func(id, title, channel string) {
		job := testutil.CreateTestJob(id, "https://youtube.com/watch?v="+id)
		repo.Create(job)
		meta := testutil.CreateTestVideoMetadata()
		meta.Title = title
		meta.Channel = channel
		meta.Tags = nil
		meta.Categories = nil
		repo.StoreMetadata(id, meta)
	}

	makeVideo("v1", "Golang Tutorial", "Tech Channel")
	makeVideo("v2", "Cooking Pasta 100%", "Food Network")
	makeVideo("v3", "Advanced golang patterns", "Tech Channel")

	query := func(search, tag string) ([]*domain.JobWithMetadata, int) {
		items, total, err := repo.GetMetadataByType("videos", domain.MetadataQuery{
			Page: 1, Limit: 20, SortBy: "created_at", Order: "desc",
			Search: search, Tag: tag,
		})
		if err != nil {
			t.Fatalf("GetMetadataByType(search=%q tag=%q) error = %v", search, tag, err)
		}
		return items, total
	}

	if _, total := query("golang", ""); total != 2 {
		t.Errorf("search 'golang' total = %d, want 2", total)
	}
	// Search must match the channel too.
	if _, total := query("food network", ""); total != 1 {
		t.Errorf("search 'food network' total = %d, want 1", total)
	}
	// LIKE wildcards in the query must be treated literally.
	if _, total := query("100%", ""); total != 1 {
		t.Errorf("search '100%%' total = %d, want 1", total)
	}
	if _, total := query("no-such-video", ""); total != 0 {
		t.Errorf("search 'no-such-video' total = %d, want 0", total)
	}

	// Tag filter: only v1 gets the user tag; auto channel tags exist too.
	if _, err := repo.AddTagsToJob("v1", []string{"watch-later"}, domain.TagSourceUser); err != nil {
		t.Fatalf("AddTagsToJob() error = %v", err)
	}
	items, total := query("", "watch-later")
	if total != 1 || len(items) != 1 {
		t.Fatalf("tag filter total = %d (items %d), want 1", total, len(items))
	}
	if items[0].Job.ID != "v1" {
		t.Errorf("tag filter returned job %s, want v1", items[0].Job.ID)
	}
	if !containsName(items[0].Tags, "watch-later") {
		t.Errorf("listing did not attach tags: %v", tagNames(items[0].Tags))
	}

	// Combined search + tag.
	if _, total := query("golang", "watch-later"); total != 1 {
		t.Errorf("combined search+tag total = %d, want 1", total)
	}

	// Auto tag from channel should match case-insensitively.
	if _, total := query("", "tech channel"); total != 2 {
		t.Errorf("tag filter 'tech channel' total = %d, want 2", total)
	}
}

func TestJobRepository_DeleteJob(t *testing.T) {
	db := testutil.CreateTestDB(t)
	defer db.Close()

	repo := NewJobRepository(db)

	video := testutil.CreateTestJob("video-1", "https://youtube.com/watch?v=video-1")
	repo.Create(video)
	repo.StoreMetadata("video-1", testutil.CreateTestVideoMetadata())

	parent := testutil.CreateTestJob("parent-1", "https://youtube.com/playlist?list=test")
	repo.Create(parent)
	repo.StoreMetadata("parent-1", testutil.CreateTestPlaylistMetadata())
	repo.AddVideoToParent("video-1", "parent-1", "playlist")

	if err := repo.DeleteJob("video-1"); err != nil {
		t.Fatalf("DeleteJob() error = %v", err)
	}

	if _, err := repo.GetByID("video-1"); err == nil {
		t.Error("job row still exists after DeleteJob()")
	}
	if count, _ := repo.CountVideos(); count != 0 {
		t.Errorf("video metadata still exists after DeleteJob(): count = %d", count)
	}
	if videos, _ := repo.GetVideosForParent("parent-1"); len(videos) != 0 {
		t.Errorf("membership still exists after DeleteJob(): %d", len(videos))
	}
	if tags, _ := repo.GetTagsForJob("video-1"); len(tags) != 0 {
		t.Errorf("tags still attached after DeleteJob(): %v", tagNames(tags))
	}

	// The parent must remain untouched.
	if _, err := repo.GetByID("parent-1"); err != nil {
		t.Errorf("parent job was deleted too: %v", err)
	}
}
