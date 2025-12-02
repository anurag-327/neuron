import fs from "fs";

const JOB_COUNT = 50;
const API_SUBMIT = "http://localhost:8080/api/v1/runner/submit";
const API_STATUS = "http://localhost:8080/api/v1/runner";

// --------------------------------------------
// Helper: POST JSON
// --------------------------------------------
async function postJSON(url, data) {
  try {
    const res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    return await res.json();
  } catch {
    return null;
  }
}

// --------------------------------------------
// Helper: GET JSON
// --------------------------------------------
async function getJSON(url) {
  try {
    const res = await fetch(url);
    return await res.json();
  } catch {
    return null;
  }
}

// --------------------------------------------
// Wait for job completion + track durations
// --------------------------------------------
async function waitForResult(jobId, submittedAt) {
  while (true) {
    const res = await getJSON(`${API_STATUS}/${jobId}/status`);

    if (!res || !res.data) {
      await new Promise((r) => setTimeout(r, 200));
      continue;
    }

    const data = res.data;

    if (["success", "failed", "canceled"].includes(data.status)) {
      const completedAt = Date.now();
      return {
        jobId,
        status: data.status,
        stdout: data.stdout,
        stderr: data.stderr,
        sandboxErrorType: data.sandboxErrorType,
        sandboxErrorMessage: data.sandboxErrorMessage,

        // Time tracking
        submittedAt,
        completedAt,
        totalDurationMs: completedAt - submittedAt,
      };
    }

    await new Promise((r) => setTimeout(r, 150));
  }
}

// --------------------------------------------
// MAIN LOAD TEST
// --------------------------------------------
(async () => {
  console.log(`🚀 Starting load test (${JOB_COUNT} jobs)…`);
  console.time("load-test");

  const submitTimes = {};  // job index → submittedAt timestamp

  // 1) Submit all jobs
  const submitPromises = [];

  for (let i = 0; i < JOB_COUNT; i++) {
    const submittedAt = Date.now();
    submitTimes[i] = submittedAt;

    submitPromises.push(
      postJSON(API_SUBMIT, {
        language: "python",
        code: "print(1)",
        input: ""
      })
    );
  }

  const submitResponses = await Promise.all(submitPromises);

  const jobIds = submitResponses
    .map((r, idx) => ({
      id: r?.data?.jobId,
      submittedAt: submitTimes[idx]
    }))
    .filter((x) => x.id);

  console.log(`📌 Submitted ${jobIds.length}/${JOB_COUNT} jobs`);

  if (jobIds.length === 0) {
    fs.writeFileSync("stress.json", JSON.stringify({ error: "No jobs submitted", submitResponses }, null, 2));
    console.log("❌ Saved debug info to stress.json");
    return;
  }

  // 2) Poll all jobs
  const statusPromises = jobIds.map((j) =>
    waitForResult(j.id, j.submittedAt)
  );

  const allResults = await Promise.all(statusPromises);

  console.timeEnd("load-test");

  // Stats
  const success = allResults.filter((r) => r.status === "success").length;
  const failed = allResults.filter((r) => r.status === "failed").length;

  console.log(`\n⏱ Timing Summary`);
  const avgTime =
    allResults.reduce((sum, r) => sum + r.totalDurationMs, 0) /
    allResults.length;

  console.log(`   Average Duration: ${avgTime.toFixed(2)} ms`);

  console.log(`\n✅ Success: ${success}`);
  console.log(`❌ Failed : ${failed}`);
  console.log(`📦 Total  : ${allResults.length}`);

  // --------------------------------------------
  // Save results to stress.json
  // --------------------------------------------
  fs.writeFileSync(
    "stress.json",
    JSON.stringify(
      {
        summary: {
          totalSubmitted: jobIds.length,
          success,
          failed,
          averageDurationMs: avgTime,
        },
        jobs: allResults,
      },
      null,
      2
    )
  );

  console.log("\n📄 Results saved to stress.json\n");
})();
