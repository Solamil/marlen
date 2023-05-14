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
	'?lang='+langCode;
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
var langCode = "";
var cookieList = document.cookie.split("; ");
var cookieLang = cookieList.find(e => e.startsWith("lang"));
if (cookieLang != undefined) {
	lang = cookieLang.split("=")[1];
	langCode = lang.split("-")[0];
}
var cookieBgColor = cookieList.find(e => e.startsWith("bgColor"));
var bgColor = "893531"
if (cookieBgColor != undefined) {
	bgColor = cookieBgColor.split("=")[1]
}

var weatherOptions = '0pq_transparency=255_background='+bgColor+'_lang='+langCode;
var weatherImgAlt = 'Current weather in ';
