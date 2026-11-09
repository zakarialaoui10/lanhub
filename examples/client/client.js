import { io } from "socket.io-client";

const socket = io("http://127.0.0.1:8080", {
  path: "/socket.io/",
  transports: ["websocket", "polling"]
});

socket.on("connect", () => {
  console.log("connected", socket.id);
  socket.emit("device:identify", {
    name: "Browser client",
    platform: "web",
    metadata: { app: "lanhub-demo" }
  });
  socket.emit("chat:message", { text: "Hello from Socket.IO client" });
});

socket.on("lanhub:ready", data => console.log("ready", data));
socket.on("chat:received", data => console.log("ack", data));
socket.on("disconnect", reason => console.log("disconnected", reason));
