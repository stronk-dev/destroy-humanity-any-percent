import crypto from "node:crypto";
import http from "node:http";

const html = `<!doctype html><html><body><main data-surface="vision_slide" aria-busy="false"><button>BEGIN ATTEMPT</button></main><script>
document.querySelector("button").addEventListener("click", () => {
  localStorage.removeItem("cloud-clicker.bootstrap-key.v1");
  localStorage.setItem("cloud-clicker.credentials.v1", JSON.stringify({present:true}));
  const socket = new WebSocket("ws://host.docker.internal:18090/connection/websocket");
  document.querySelector("main").outerHTML = '<main data-surface="desk" aria-busy="false"><p>You are visitor #1</p><button>Fix Computer</button></main>';
  document.querySelector("button").addEventListener("click", async () => {
    document.querySelector("main").setAttribute("aria-busy", "true");
    await fetch("/api/v1/intents", {method:"POST", headers:{"content-type":"application/json"}, body:"{}"});
    document.querySelector("main").setAttribute("aria-busy", "false");
  });
});
</script></body></html>`;

const server = http.createServer((request, response) => {
  if (request.method === "GET" && request.url === "/") {
    response.writeHead(200, { "content-type": "text/html", "content-length": Buffer.byteLength(html) });
    response.end(html);
    return;
  }
  if (request.method === "POST" && request.url === "/api/v1/intents") {
    request.resume();
    response.writeHead(200, { "content-type": "application/json" });
    response.end('{"outcome":"applied"}\n');
    return;
  }
  response.writeHead(404);
  response.end();
});

server.on("upgrade", (request, socket) => {
  if (request.url !== "/connection/websocket" || typeof request.headers["sec-websocket-key"] !== "string") {
    socket.destroy();
    return;
  }
  const accept = crypto.createHash("sha1").update(request.headers["sec-websocket-key"] + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11").digest("base64");
  socket.write("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n");
});

server.listen(18090, "0.0.0.0", () => console.log("deployment browser fixture ready"));
