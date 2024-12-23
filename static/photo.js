(() => {
  const dropArea = document.getElementById("drop-area");
  const inputFile = document.getElementById("input-file");
  const imageView = document.getElementById("image-view");

  inputFile.addEventListener('change', changeImage)

  function changeImage() {
    const firstFile = inputFile.files[0];
    const imageLink = URL.createObjectURL(firstFile);
    imageView.style.backgroundImage = `url(${imageLink})`;
    imageView.textContent = "";
    showPreview();
  };

  dropArea.addEventListener('drop', (ev) => {
    ev.preventDefault();
    inputFile.files = ev.dataTransfer.files;
    changeImage();
  });

  dropArea.addEventListener('dragover', (ev) => {
    ev.preventDefault();
  });

  const video = document.getElementById("video");
  const canvas = document.getElementById("canvas");
  const captureButton = document.getElementById("capture");
  const webcamDiv = document.getElementById("webcam");
  const previewDiv = document.getElementById("preview");

  function showLive() {
    previewDiv.style.display = "none";
    webcamDiv.style.display = "block"
  }

  function showPreview() {
    previewDiv.style.display = "block";
    webcamDiv.style.display = "none";
  }

  navigator.mediaDevices.getUserMedia({ video: true })
    .then((stream) => {
      showLive();
      video.srcObject = stream;
    })
    .catch((err) => {
      console.info("Access to the camera is not possible", err)
      showPreview();
    });

  captureButton.addEventListener("click", (event) => {
    event.preventDefault();
    const context = canvas.getContext("2d");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    context.drawImage(video, 0, 0, canvas.width, canvas.height);

    canvas.toBlob((blob) => {
      const file = new File([blob], "webcam.png", { type: "image/png" });
      const dataTransfer = new DataTransfer();
      dataTransfer.items.add(file);
      inputFile.files = dataTransfer.files;
      changeImage();
    }, 'image/png');
  });
})();
