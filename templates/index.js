var ws = null;

function myWebsocketStart() {
  var ws = new WebSocket("ws://localhost:3000/websocket");
  ws.onopen = function() {
    // Web Socket is connected, send data using send()
    ws.send(JSON.stringify({ type: "connect" }));
    var myTextArea = document.getElementById("textarea1");
    myTextArea.value = "First message sent";
    console.log("websocket opened");
  };

  ws.onmessage = function(evt) {
    console.log("websocket event recieved: ", evt);
    // var myTextArea = document.getElementById("textarea1");
    // myTextArea.value = myTextArea.value + "\n" + evt.data;
    // ws.send(JSON.stringify({ type: "up" }));
    // if (evt.data == "pong") {
    //   setTimeout(function() {
    //     ws.send(JSON.stringify({ type: "dig" }));
    //   }, 2000);
    // }
  };

  ws.onclose = function() {
    console.log("connection closed");
    var myTextArea = document.getElementById("textarea1");
    myTextArea.value = myTextArea.value + "\n" + "Connection closed";
  };
}

function websocketClose() {
  console.log("time to close websocket");
  if (ws !== null) {
    ws.close();
    ws = null;
  }
  console.log("websocket: ", ws);
}
