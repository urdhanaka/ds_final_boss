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

  // accessClusterButton = document.getElementById("access-button")
  // accessClusterButton.style.visibility = "visible"

  document.getElementById("wait").appendChild(document.createTextNode("creating cluster, please wait..."));

  try {
    const response = await fetch("http://localhost:3000/create_cluster", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data)
    });

    const result = await response.json();

    // streamLog()
    let tokenContent = document.createTextNode(result.data.dashboard_token);
    document.getElementById("token").appendChild(tokenContent);

    // set the session key
    // sessionKey = result;

    // enable the button to access the cluster
    // accessClusterButton = document.getElementById("access-button")
    // accessClusterButton.style.visibility = "visible"
  } catch (error) {
    console.error('Error:', error.message);
    document.getElementById("wait").appendChild(document.createTextNode("error creating cluster: ", error.message));
  } finally {
    document.getElementById("wait").appendChild(document.createTextNode("done"));
  }
}

function accessCluster() {
  return;
}

async function streamLog() {
  const logTag = document.getElementById("log")

  let fullWebsocketUrl = "ws://localhost:3000/status/" + clusterName;

  const socket = new WebSocket(fullWebsocketUrl);

  socket.addEventListener("message", (event) => {
    const logLine = event.data;
    logTag.textContent += logLine + "\n"
  });

  socket.addEventListener("close", () => {
    logTag.textContent += "--END OF LOG--\n"
  });

  socket.addEventListener("error", (error) => {
    logTag.textContent += `--ERROR OCCURED--\n${error.message}`;
  });
}
