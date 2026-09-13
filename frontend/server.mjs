import { createReadStream, existsSync, statSync } from "node:fs";
import { createServer } from "node:http";
import { extname, join, normalize } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(fileURLToPath(new URL(".", import.meta.url)), "build", "client");
const port = Number(process.env.PORT || 3000);

const contentTypes = new Map([
  [".css", "text/css; charset=utf-8"],
  [".html", "text/html; charset=utf-8"],
  [".ico", "image/x-icon"],
  [".js", "text/javascript; charset=utf-8"],
  [".json", "application/json; charset=utf-8"],
  [".png", "image/png"],
  [".svg", "image/svg+xml"],
  [".webp", "image/webp"],
  [".woff2", "font/woff2"],
]);

function resolvePath(url) {
  const pathname = new URL(url, "http://localhost").pathname;
  const normalized = normalize(decodeURIComponent(pathname)).replace(/^([/\\])+/, "");
  const requested = join(root, normalized);
  if (requested.startsWith(root) && existsSync(requested) && statSync(requested).isFile()) {
    return requested;
  }
  return join(root, "index.html");
}

createServer((request, response) => {
  const filePath = resolvePath(request.url || "/");
  response.setHeader(
    "Content-Type",
    contentTypes.get(extname(filePath)) || "application/octet-stream"
  );
  createReadStream(filePath)
    .on("error", () => {
      response.statusCode = 404;
      response.end("Not found");
    })
    .pipe(response);
}).listen(port, "0.0.0.0", () => {
  console.log(`Serving Sabeel frontend on http://0.0.0.0:${port}`);
});
