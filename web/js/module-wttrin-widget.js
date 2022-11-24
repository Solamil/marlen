function changeWeather () {
	var input = document.getElementById('weather-input');
	var img = document.getElementById('weather-img');
	var wttrLink = document.getElementById("wttr-link")
	img.src =
	'https://wttr.in/' + input.value +
	'_' + weatherOptions
	+ '.png';
	img.alt = weatherImgAlt + input.value;
	wttrLink.href = 'https://wttr.in/' + input.value +
	'?lang='+langOption;
	input.value = '';
}

document.getElementById('weather-btn').onclick = function () {
	changeWeather();
}

document.getElementById('weather-form').onkeypress = function(event) {
	if (event.keyCode === 13) {
	  event.preventDefault();
	  changeWeather();
	}
}
var langOption = "";
var cookieList = document.cookie.split("; ");
var cookieLang = cookieList.find(e => e.startsWith("lang"));
if (cookieLang != undefined) {
	lang = cookieLang.split("=")[1];
	langOption = lang.split("-")[0];
}

var weatherOptions = '0pq_transparency=255_background=893531_lang=en';
var weatherImgAlt = 'Current weather in ';
