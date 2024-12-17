const dropArea = document.getElementById("drop-area");
const inputFile = document.getElementById("input-file");
const imageView = document.getElementById("image-view");

inputFile.addEventListener('change', changeImage)

function changeImage() {
  const firstFile = inputFile.files[0];
  const imageLink = URL.createObjectURL(firstFile);
  imageView.style.backgroundImage = `url(${imageLink})`;
  imageView.textContent = "";
}

dropArea.addEventListener('drop', (ev) => {
  ev.preventDefault();
  inputFile.files = ev.dataTransfer.files;
  changeImage();
})

dropArea.addEventListener('dragover', (ev) => {
  ev.preventDefault();
})
