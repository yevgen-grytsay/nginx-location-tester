function initApp({ wsUrl, nginxUrl, fetchViaProxy }) {
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
      const ls = new LogSequence(data);
      let textContent = "";

      if (typeof data === "object") {
        textContent = ls.getSimpleLines().join("<br>");
      } else {
        textContent = JSON.stringify(data, null, "  ");
      }

      // const prettyJson = JSON.stringify(data, null, "  ");
      var d = document.createElement("div");
      // d.innerHTML = "<pre>" + prettyJson + "</pre>";
      d.innerHTML = "<pre>" + textContent + "</pre>";
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

    document.querySelectorAll("[data-asset-file]").forEach((el) => {
      el.addEventListener("click", (evt) => {
        const fileName = evt.target.dataset.assetFile;

        if (fetchViaProxy) {
          const url = new URL("/fetch", document.baseURI);
          url.searchParams.set("file", fileName);
          fetch(url);
        } else {
          const url = new URL(fileName, nginxUrl);
          fetch(url, { mode: "no-cors" });
        }
      });
    });

    document
      .querySelector(".custom-url button")
      .addEventListener("click", (evt) => {
        const fileName = document.querySelector(".custom-url input").value;

        if (fetchViaProxy) {
          const url = new URL("/fetch", document.baseURI);
          url.searchParams.set("file", fileName);
          fetch(url);
        } else {
          const url = new URL(fileName, nginxUrl);
          fetch(url, { mode: "no-cors" });
        }
      });
  });

  /**
   * @typedef {object} LogLine
   * @property {string} Message
   */
  /**
   * @typedef {Object} LogSequenceResponse
   * @property {LogLine[]} Lines
   * @property {object} RequestFullId
   */

  class LogSequence {
    /**
     * @param {LogSequenceResponse} data
     */
    constructor(data) {
      this.data = data;
    }

    getSimpleLines() {
      return this.data.Lines.map((line) => {
        return line.Message;
      });
    }
  }
}
