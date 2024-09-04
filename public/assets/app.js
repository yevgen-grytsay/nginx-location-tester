function initApp({ wsUrl }) {
  window.addEventListener("load", function (evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function (message) {
      var d = document.createElement("div");
      d.innerHTML = message.replaceAll(/\n+/g, "\n").replaceAll("\n", "<br>");
      output.appendChild(d);
      output.scroll(0, output.scrollHeight);
    };

    const printResponse = function (body) {
      const data = JSON.parse(body);
      const prettyJson = JSON.stringify(data, null, "  ");
      var d = document.createElement("div");
      d.innerHTML = "<pre>" + prettyJson + "</pre>";
      output.appendChild(d);
      output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function (evt) {
      if (ws) {
        return false;
      }
      ws = new WebSocket(wsUrl);
      ws.onopen = function (evt) {
        print("OPEN");
      };
      ws.onclose = function (evt) {
        print("CLOSE");
        ws = null;
      };
      ws.onmessage = function (evt) {
        printResponse(evt.data);
      };
      ws.onerror = function (evt) {
        print("ERROR: " + evt.data);
      };
      return false;
    };

    document.getElementById("send").onclick = function (evt) {
      if (!ws) {
        return false;
      }
      print("SEND: " + input.value);
      ws.send(input.value);
      return false;
    };

    document.getElementById("close").onclick = function (evt) {
      if (!ws) {
        return false;
      }
      ws.close();
      return false;
    };
  });
}
