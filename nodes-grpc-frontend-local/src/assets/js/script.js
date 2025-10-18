// will be used later
var sessionKey = "";

async function sendRequirementJson() {
  let name = document.getElementById("name").value;
  let vcpu = document.getElementById("vcpu").value;
  let memory = document.getElementById("memory").value;
  let storage = document.getElementById("storage").value;
  let node_size = document.getElementById("node_size").value;

  const data = {
    name: name,
    vcpu: vcpu,
    memory: memory,
    storage: storage,
    node_size: node_size,
  }

  const fetchPromise = fetch("http://localhost:3000/create_cluster", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data)
  });

  document.getElementById("wait").appendChild(document.createTextNode("creating cluster, please wait..."));

  // setTimeout(() => {
  //   streamLog(name)
  // }, 1000);

  try {
    const response = await fetchPromise;
    const result = await response.json();

    // streamLog()
    let tokenContent = document.createTextNode(result.data.dashboard_token);
    document.getElementById("token").appendChild(tokenContent);

    document.getElementById("accessClusterButton").onclick = function() { accessCluster("https://" + result.data.master_ip_address + ":8443") };
    document.getElementById("accessClusterButton").classList.remove("invisible")
    document.getElementById("accessClusterButton").style.visibility = "visible"

  } catch (error) {
    console.error('Error:', error.message);
    document.getElementById("wait").appendChild(document.createTextNode("error creating cluster: ", error.message));
  } finally {
    document.getElementById("wait").appendChild(document.createTextNode("done"));
  }
}

function accessCluster(url) {
  window.open(url, '_blank').focus()
}

function streamLog(clusterName) {
  const logTag = document.getElementById("kernel-log")

  let fullWebsocketUrl = "ws://localhost:3000/ws/stream_logs/" + clusterName;

  const socket = new WebSocket(fullWebsocketUrl);

  socket.addEventListener("message", (event) => {
    const logLine = event.data;

    var pTag = document.createElement("p")
    pTag.appendChild(document.createTextNode(logLine))

    logTag.appendChild(pTag)
  });

  socket.addEventListener("close", () => {
    var pTag = document.createElement("p")
    pTag.appendChild(document.createTextNode("--END OF LOG--"))
    logTag.appendChild(pTag)
  });

  socket.addEventListener("error", (error) => {
    var pTag = document.createElement("p")
    pTag.appendChild(document.createTextNode(`--ERROR OCCURED ${error.message}--`))
    logTag.appendChild(pTag)
  });
}
