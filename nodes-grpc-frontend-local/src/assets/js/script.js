// will be used later
var sessionKey = "";

async function sendRequirementJson() {
  let name = document.getElementById("name").value;
  let vcpu = document.getElementById("vcpu").value;
  let memory = document.getElementById("memory").value;
  let storage = document.getElementById("storage").value;

  const data = {
    name: name,
    vcpu: vcpu,
    memory: memory,
    storage: storage,
  }

  // accessClusterButton = document.getElementById("access-button")
  // accessClusterButton.style.visibility = "visible"

  try {
    const response = await fetch("http://localhost:3000/create_cluster", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data)
    });

    const result = await response.json();
    console.log('Success:', result);

    // set the session key
    // sessionKey = result;

    // enable the button to access the cluster
    // accessClusterButton = document.getElementById("access-button")
    // accessClusterButton.style.visibility = "visible"
  } catch (error) {
    console.error('Error:', error.message);
  }
}

function openDashboard() {
  return;
}

async function streamLog(clusterName) {
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
