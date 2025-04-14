async function sendRequirementJson() {
  let vcpu = document.getElementById("vcpu").value;
  let memory = document.getElementById("memory").value;

  const data = {
    vcpu: vcpu,
    memory: memory,
  }

  console.log(data)

  // try {
  //   const response = await fetch("http://localhost:3000/create_cluster", {
  //     method: "POST",
  //     headers: {
  //       "Content-Type": "application/json",
  //     },
  //     body: JSON.stringify(data)
  //   });
  //
  //   console.log("here")
  //
  //   const result = await response.json();
  //   console.log('Success:', result);
  //
  // } catch (error) {
  //   console.error('Error:', error.message);
  // }
}
