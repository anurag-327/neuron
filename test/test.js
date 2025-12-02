import fs from "fs";

// ------------------------------------------------------------------------------------
// API ENDPOINTS
// ------------------------------------------------------------------------------------
const API_SUBMIT = "http://localhost:8080/api/v1/runner/submit";
const API_STATUS = "http://localhost:8080/api/v1/runner";

// ------------------------------------------------------------------------------------
// FETCH HELPERS
// ------------------------------------------------------------------------------------

async function postJSON(url, data) {
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data)
  });

  return res.json();
}

async function getJSON(url) {
  const res = await fetch(url);
  return res.json();
}

// ------------------------------------------------------------------------------------
// TEST CASES (WITH EXPECTED RESULTS)
// ------------------------------------------------------------------------------------

const tests = [
  // ---------------- CPP -----------------
  {
    name: "CPP: Hello World",
    expected: "Hello CPP",
    body: {
      code: '#include <iostream>\nint main(){ std::cout<<"Hello CPP"; }',
      input: "",
      language: "cpp"
    }
  },
  {
    name: "CPP: Compilation Error",
    expectedError: "CompilationError",
    body: {
      code: '#include <iostream>\nint main(){ std::cout << x; }',
      input: "",
      language: "cpp"
    }
  },

  // ---------------- PYTHON -----------------
  {
    name: "Python: Runtime Error (division by zero)",
    expectedError: "RuntimeError",
    body: {
      code: "x = 5\ny = 0\nprint(x / y)",
      input: "",
      language: "python"
    }
  },
  {
    name: "Python: Heavy Loop",
    expected: "49999995000000",
    body: {
      code: "s=0\nfor i in range(10_000_000): s+=i\nprint(s)",
      input: "",
      language: "python"
    }
  },

  // ---------------- JAVASCRIPT -----------------
  {
    name: "JS: BFS Traversal + Heavy Load",
    expectedContains: "Reachable:",
    body: {
      code: `
function bfs(graph, start) {
  let visited = new Set();
  let queue = [start];
  visited.add(start);

  while (queue.length > 0) {
    let node = queue.shift();
    for (let nei of graph[node] || []) {
      if (!visited.has(nei)) {
        visited.add(nei);
        queue.push(nei);
      }
    }
  }
  return visited.size;
}

const N = 2000;
let graph = {};
for (let i = 1; i <= N; i++) graph[i] = [];

for (let i = 1; i <= N; i++) {
  for (let j = 0; j < 3; j++) {
    let to = Math.floor(Math.random() * N) + 1;
    if (to !== i) graph[i].push(to);
  }
}

let reachable = bfs(graph, 1);
console.log("Reachable:", reachable);
`,
      input: "",
      language: "js"
    }
  },
  {
    name: "JS: ReferenceError",
    expectedError: "RuntimeError",
    body: {
      code: "console.log(x + 1);",
      input: "",
      language: "js"
    }
  },

  // ---------------- GO -----------------
  {
    name: "Go: Panic Test",
    expectedError: "RuntimeError",
    body: {
      code: `package main
import "fmt"

func main() {
    var x []int
    fmt.Println(x[10]) // panic
}`,
      input: "",
      language: "go"
    }
  },
  {
    name: "Go: Normal Program",
    expected: "55",
    body: {
      code: `package main
import "fmt"

func main() {
  sum := 0
  for i:=1; i<=10; i++ { sum+=i }
  fmt.Println(sum)
}`,
      input: "",
      language: "go"
    }
  },

  // ---------------- JAVA -----------------
  {
    name: "Java: Dijkstra Shortest Path",
    expectedContains: "0",
    body: {
      code: `import java.util.*;

public class Main {
  static class P { int v; long w; P(int v,long w){this.v=v;this.w=w;} }
  public static void main(String[] a){
    Scanner sc=new Scanner(System.in);
    int N=sc.nextInt(), M=sc.nextInt();
    List<List<P>> g = new ArrayList<>();
    for(int i=0;i<=N;i++) g.add(new ArrayList<>());
    for(int i=0;i<M;i++){
      int u=sc.nextInt(), v=sc.nextInt(); long w=sc.nextLong();
      g.get(u).add(new P(v,w));
      g.get(v).add(new P(u,w));
    }
    long[] d=new long[N+1];
    Arrays.fill(d, Long.MAX_VALUE);
    PriorityQueue<long[]> pq=new PriorityQueue<>(Comparator.comparingLong(x->x[1]));
    d[1]=0;
    pq.add(new long[]{1,0});
    while(!pq.isEmpty()){
      long[] cur=pq.poll();
      int node=(int)cur[0];
      if(cur[1]!=d[node]) continue;
      for(P nx:g.get(node)){
        if(d[nx.v]>cur[1]+nx.w){
          d[nx.v]=cur[1]+nx.w;
          pq.add(new long[]{nx.v,d[nx.v]});
        }
      }
    }
    for(int i=1;i<=N;i++) System.out.print(d[i]+" ");
  }
}`,
      input: "5 6\n1 2 3\n1 3 4\n2 3 2\n2 4 7\n3 5 1\n4 5 2\n",
      language: "java"
    }
  }
];

// ------------------------------------------------------------------------------------
// SUBMIT JOB
// ------------------------------------------------------------------------------------

async function submitJob(body) {
  const res = await postJSON(API_SUBMIT, body);
  return res.data.jobId;
}

// ------------------------------------------------------------------------------------
// WAIT FOR JOB COMPLETION
// ------------------------------------------------------------------------------------

async function waitForResult(jobId) {
  while (true) {
    const res = await getJSON(`${API_STATUS}/${jobId}/status`);
    const data = res.data;

    if (["success", "failed", "canceled"].includes(data.status)) {
      return data;
    }

    await new Promise(r => setTimeout(r, 200));
  }
}

// ------------------------------------------------------------------------------------
// PASS/FAIL VALIDATION
// ------------------------------------------------------------------------------------

function checkTest(test, result) {
  const stdout = result.stdout.trim();
  const errType = result.sandboxErrorType;

  if (test.expected !== undefined) {
    return stdout === test.expected;
  }

  if (test.expectedContains !== undefined) {
    return stdout.includes(test.expectedContains);
  }

  if (test.expectedError !== undefined) {
    return errType === test.expectedError;
  }

  return false; // default
}

// ------------------------------------------------------------------------------------
// MAIN TEST RUNNER
// ------------------------------------------------------------------------------------

async function runTests() {
  const results = [];

  for (const t of tests) {
    console.log(`\n🚀 Running test: ${t.name}`);
    const start = Date.now();

    try {
      const jobId = await submitJob(t.body);
      console.log(`➡️ Submitted Job ID: ${jobId}`);

      const result = await waitForResult(jobId);
      const end = Date.now();

      const passed = checkTest(t, result);

      results.push({
        name: t.name,
        jobId,
        expected: t.expected || t.expectedContains || t.expectedError,
        actual: {
          stdout: result.stdout,
          stderr: result.stderr,
          errorType: result.sandboxErrorType,
          errorMessage: result.sandboxErrorMessage
        },
        status: passed ? "PASS" : "FAIL",
        durationMs: end - start
      });

      console.log(passed ? `✅ PASS: ${t.name}` : `❌ FAIL: ${t.name}`);

    } catch (err) {
      results.push({
        name: t.name,
        error: err.message,
        status: "ERROR"
      });
      console.log(`❌ ERROR: ${t.name}`);
    }
  }

  fs.mkdirSync("./test", { recursive: true });
  fs.writeFileSync("./test/results.json", JSON.stringify(results, null, 2));

  console.log("\n📄 Full results saved at: ./test/results.json\n");
}

runTests();
