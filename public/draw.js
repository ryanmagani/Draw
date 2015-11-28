(function()
{
  var canvas = document.getElementsByTagName('canvas')[0];
  var ctx = canvas.getContext("2d");

  var mouseX = 0;
  var mouseY = 0;
  var prevMouseX = 0;
  var prevMouseY = 0;

  var isDrawer = true;

  var textbox = document.getElementsByTagName('input')[0];
  var submitBtn = document.getElementsByTagName('input')[1];

  var xhttp = new XMLHttpRequest();

  canvas.addEventListener('mousemove', function(e) {
    mousePos(e);
  }, false);

  canvas.addEventListener('mousedown', function(e) {
    canvas.addEventListener('mousemove', draw, false);
  }, false);

  canvas.addEventListener('mouseup', function(e) {
    canvas.removeEventListener('mousemove', draw, false);
  }, false);

  submitBtn.addEventListener('click', function(e) {
    guess();
  });

  var mousePos = function(e)
  {
    prevMouseX = mouseX;
    prevMouseY = mouseY;
    mouseX = e.clientX - canvas.offsetLeft;
    mouseY = e.clientY - canvas.offsetTop;
  }

  var draw = function()
  {
    if (!isDrawer)
      return;

    ctx.beginPath();
    ctx.moveTo(prevMouseX, prevMouseY)
    ctx.lineTo(mouseX, mouseY);
    ctx.stroke();
    ctx.closePath();
    sendToServer("/draw", "" + mouseX + "," + mouseY);
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
