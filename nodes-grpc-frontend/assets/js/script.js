async function sendRequirementJson() {
  let vcpu = document.getElementById("vcpu").value;
  let memory = document.getElementById("memory").value;

  const data = {
    vcpu: vcpu,
    memory: memory,
  }

  try {
    const response = await fetch("http://localhost:3000/create-cluster", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data)
    });

    if (response.ok) {
      const result = await response.json();
      console.log('Success:', result);
    } else {
      console.error('Error:', response.status, response.statusText);
    }
  } catch (error) {
    console.error('Error:', error.message);
  }
}
