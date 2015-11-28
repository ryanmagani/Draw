(function()
{
  var canvas = document.getElementsByTagName('canvas')[0];
  var ctx = canvas.getContext("2d");

  var mouseX = 0;
  var mouseY = 0;
  var prevMouseX = 0;
  var prevMouseY = 0;

  canvas.addEventListener('mousemove', function(e) {
    mousePos(e);
  }, false);

  canvas.addEventListener('mousedown', function(e) {
    canvas.addEventListener('mousemove', draw, false);
  }, false);

  canvas.addEventListener('mouseup', function(e) {
    canvas.removeEventListener('mousemove', draw, false);
  }, false);

  var mousePos = function(e)
  {
    prevMouseX = mouseX;
    prevMouseY = mouseY;
    mouseX = e.clientX - canvas.offsetLeft;
    mouseY = e.clientY - canvas.offsetTop;
  }

  var draw = function()
  {
    ctx.beginPath();
    ctx.moveTo(prevMouseX, prevMouseY)
    ctx.lineTo(mouseX, mouseY);
    ctx.stroke();
    ctx.closePath();
  }

}());
