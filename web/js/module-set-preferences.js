function changeLocation() {
	var input = document.getElementById("location-input")
	var date = new Date(Date.now())
	date.setFullYear(date.getFullYear()+1)
	document.cookie = "location="+input.value+"; expires="+date.toUTCString()+"; SameSite=Lax; HttpOnly:true; Secure: true"
}
document.getElementById("location-btn").onclick = function () {
	changeLocation()
}
