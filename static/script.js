console.log("Hello World");
document
  .getElementById("registrationForm")
  .addEventListener("submit", function (e) {
    e.preventDefault(); // Prevent page reload

    const first_name = document.getElementById("first_name").value;
    const last_name = document.getElementById("last_name").value;
    const username = document.getElementById("nickname").value;
    const age = document.getElementById("age").value;
    // const gender = document.querySelector('input[name="gender"]:checked').value;
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    if (
      !username ||
      !password ||
      // !gender ||
      !first_name ||
      !last_name ||
      !age ||
      !email
    ) {
      alert("All fields are required");
      return;
    }

    handleRegistration({
      first_name,
      last_name,
      username,
      age,
      // gender,
      email,
      password,
    });
  });

function handleRegistration(data) {
  console.log("User registered:", data);

  // Simulate success response
  setTimeout(() => {
    alert("Registration successful! Redirecting to login...");
    showLoginPage();
  }, 1000);
}

// function showLoginPage() {
//   document.getElementById("registrationView").classList.remove("active");
//   document.getElementById("loginView").classList.add("active");
// }
