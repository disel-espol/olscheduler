#!/usr/bin/node

const http = require("http");

const delay = process.argv[2];
const port = process.argv[3];

const logRequest = req => {
  const time = new Date().toISOString();
  const requestData = `${req.method} ${req.url}`;
  console.log(`[${time}] recieved request ${requestData}`);
};

const sendResponse = res => {
  res.writeHead(200, { "Content-Type": "text/plain" });
  res.end("Request handled by worker at " + port);
};

const requestHandler =
  delay === 0
    ? (req, res) => {
        logRequest(req);
        sendResponse(res);
      }
    : (req, res) =>
        setTimeout(() => {
          logRequest(req);
          sendResponse(res);
        }, delay * 1000);

const server = http.createServer(requestHandler);
server.listen(port, "localhost");
