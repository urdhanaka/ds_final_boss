// will be used later
var sessionKey = "";

async function sendRequirementJson() {
  let vcpu = document.getElementById("vcpu").value;
  let memory = document.getElementById("memory").value;

  const data = {
    vcpu: vcpu,
    memory: memory,
  }

  accessClusterButton = document.getElementById("access-button")
  accessClusterButton.style.visibility = "visible"

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
    sessionKey = result;

    // enable the button to access the cluster
    accessClusterButton = document.getElementById("access-button")
    accessClusterButton.style.visibility = "visible"
  } catch (error) {
    console.error('Error:', error.message);
  }
}

function openDashboard() {
  return;
}
