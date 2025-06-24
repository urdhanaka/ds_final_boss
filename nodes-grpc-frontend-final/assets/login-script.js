async function handleLogin(event) {
  event.preventDefault();

  // user data
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;

  const data = {
    "email": email,
    "password": password,
  };

  console.log("here")

  try {
    const response = await fetch("http://localhost:8000/api/users/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify(data)
    })

    const res = await response.json()

    if (response.ok) {
      alert("login successful");



      setTimeout(() => {
        window.location.href = "/dashboard";
      }, 1000)
    }
  } catch (error) {
    alert("error" + error)
  }
}
