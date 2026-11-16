const http = require("http");

const PORT = 8017;
let jammModeServeState = 0; // 0: off, 1: on, 2: auto

const server = http.createServer((req, res) => {
  // Set JSON headers
  res.setHeader("Content-Type", "application/json");

  // Handle standard GET requests for the two endpoints
  if (req.method === "GET" && req.url === "/jamm_state") {
    res.writeHead(200);
    return res.end(
      JSON.stringify({
        success: true,
        name: "JammerMock",
        firmware: "0.1.0",
        board: "MockBoard",
        desktop_mode: jammModeServeState,
        jamm_900: true,
        jamm_1500: true,
        jamm_2400: true,
        jamm_5200: false,
        jamm_5800: true,
      }),
    );
  }

  if (req.method === "GET" && req.url === "/manualon") {
    jammModeServeState = 1;
    res.writeHead(200);

    console.log("Jammer mode set to ON");
    return res.end(
      JSON.stringify({
        success: true,
        mode: jammModeServeState,
      }),
    );
  }

  if (req.method === "GET" && req.url === "/manualoff") {
    jammModeServeState = 0;
    res.writeHead(200);

    console.log("Jammer mode set to OFF");
    return res.end(
      JSON.stringify({
        success: true,
        mode: jammModeServeState,
      }),
    );
  }

  if (req.method === "GET" && req.url === "/auto") {
    jammModeServeState = 2;
    res.writeHead(200);

    console.log("Jammer mode set to AUTO");
    return res.end(
      JSON.stringify({
        success: true,
        mode: jammModeServeState,
      }),
    );
  }

  if (req.method === "GET" && req.url === "/alert") {
    res.writeHead(200);
    console.log("Jammer alert triggered");
    return res.end(
      JSON.stringify({
        success: true,
        message: "Alert triggered",
      }),
    );
  }

  // Fallback for any other route or HTTP method
  res.writeHead(404);
  return res.end(
    JSON.stringify({
      success: false,
      msg: "Route not found",
    }),
  );
});

server.listen(PORT, () => {
  console.log(`Server running at http://localhost:${PORT}/`);
});
