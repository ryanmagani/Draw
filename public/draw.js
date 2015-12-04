(function()
	{
		// client identifier
//		var guid = getId();
		var canvas = document.getElementById('canvas');
		var ctx = canvas.getContext("2d");

		var color = "black";
		var saveColor = "black";

		var mouseX = 0;
		var mouseY = 0;
		var prevMouseX = 0;
		var prevMouseY = 0;

		var isDrawer = false;

		var eraseCheck = document.getElementById('erase');
		var clearBtn = document.getElementById('clearBtn');

		var textbox = document.getElementById('guessText');
		var guessBtn = document.getElementById('guessBtn');

		var xhttp = new XMLHttpRequest();

		var ws = new WebSocket("ws://localhost:7777/join");

		var read = function(event)
		{
			isDrawer = JSON.parse(event.data);
		}

		ws.onmessage = read;


		window.addEventListener('beforeunload', function (e) {
			quit();
		}, false);

		canvas.addEventListener('mousemove', function(e) {
			mousePos(e);
		}, false);

		canvas.addEventListener('mousedown', function(e) {
			canvas.addEventListener('mousemove', draw, false);
		}, false);

		canvas.addEventListener('mouseup', function(e) {
			canvas.removeEventListener('mousemove', draw, false);
		}, false);

		clearBtn.addEventListener('click', function(e) {
			clear();
		});

		guessBtn.addEventListener('click', function(e) {
			guess();
		});

		eraseCheck.addEventListener('click', function(e) {
			toggleEraser();
		});


/*		function getId() {
			 getFromServer(""). ;
		}*/

		function mousePos(e)
		{
			prevMouseX = mouseX;
			prevMouseY = mouseY;
			mouseX = e.clientX - canvas.offsetLeft;
			mouseY = e.clientY - canvas.offsetTop;
		}

		function draw()
		{
			if (!isDrawer)
				return;

			ctx.strokeStyle = color;
			ctx.beginPath();
			ctx.moveTo(prevMouseX, prevMouseY)
				ctx.lineTo(mouseX, mouseY);
			ctx.stroke();
			ctx.closePath();
			sendToServer("/draw", "" + mouseX + "," + mouseY + "," + color);
		}

		function toggleEraser()
		{
			if (eraseCheck.checked)
			{
				saveColor = color;
				color = "white";
			}

			else
			{
				color = saveColor;
			}
		}

		function clear()
		{
			ctx.clearRect(0, 0, canvas.width, canvas.height);
		}

		function quit()
		{
			sendToServer("/quit", "1");
		}

		var guess = function()
		{
			// if (isDrawer)
				//   return;

			sendToServer("/guess", textbox.value);
		}

		function sendToServer(url, data)
		{
			xhttp.open("POST", url, true);
			xhttp.send(data);
		}

	}());
