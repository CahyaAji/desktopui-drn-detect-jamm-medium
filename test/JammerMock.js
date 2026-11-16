const http = require("http");

const PORT = 8017;
let jammModeServeState = 0; // 0: off, 1: on, 2: auto
let jammBands = {
  jamm_900: true,
  jamm_1500: true,
  jamm_2400: true,
  jamm_5200: false,
  jamm_5800: true,
};

const server = http.createServer((req, res) => {
  // Set JSON headers
  res.setHeader("Content-Type", "application/json");

  // Handle standard GET requests for the two endpoints
  if (req.method === "GET" && req.url === "/state") {
    res.writeHead(200);
    return res.end(
      JSON.stringify({
        success: true,
        name: "JammerMock",
        firmware: "0.1.0",
        board: "MockBoard",
        desktop_mode: jammModeServeState,
        jamm_900: jammBands.jamm_900 ? 1 : 0,
        jamm_1500: jammBands.jamm_1500 ? 1 : 0,
        jamm_2400: jammBands.jamm_2400 ? 1 : 0,
        jamm_5200: jammBands.jamm_5200 ? 1 : 0,
        jamm_5800: jammBands.jamm_5800 ? 1 : 0,
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

  if (req.method === "POST" && req.url === "/deskcmd") {
    let body = "";
    req.on("data", (chunk) => {
      body += chunk;
    });
    req.on("end", () => {
      try {
        const payload = JSON.parse(body);
        const { desktop_mode, ...bands } = payload;
        jammBands = { ...jammBands, ...bands };
        if (typeof desktop_mode === "number") {
          jammModeServeState = desktop_mode;
        }
      } catch (error) {
        console.error("Invalid JSON body:", error);
      }
      console.log("Jammer bands updated via POST /deskcmd");
      res.writeHead(200);
      res.end(
        JSON.stringify({
          success: true,
          desktop_mode: jammModeServeState,
        }),
      );
    });
    return;
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
