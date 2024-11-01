'use client'

import {UrlInput} from "@/components/url-input";
import JobProgress from "@/components/job-progress";

export default function Home() {


    return (
    <div className="flex flex-col w-full min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col w-full">
          <UrlInput />
          <JobProgress />
      </main>
    </div>
  );
}
