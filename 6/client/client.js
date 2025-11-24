import WebSocket from "ws";
import fetch from "node-fetch";

const clientID = process.argv[2] || "default";

async function runJob(input) {
  // make a http post request and submit the job
  const res = await fetch("http://localhost:8080/submit", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ input }),
  });
  const job = await res.json();
  console.log(`[Client ${clientID}] submitted job:`, job);

  // new websocket connection

  const ws = new WebSocket(`ws://localhost:8080/ws?job_id=${job.id}`);

  ws.on("open", () => console.log("Waiting for job completion..."));
  ws.on("message", (msg) => {
    const data = JSON.parse(msg);
    console.log(`[Client ${clientID}] Job update received:`, data);
    if (data.status === "done") {
      console.log(
        `[Done!!!! Client ${clientID}] Job ${data.job_id} completed with result: ${data.result}`
      );
      ws.close();
    }
  });
}

runJob("hello world");
