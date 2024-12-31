import { BarcodeDetector } from "https://fastly.jsdelivr.net/npm/barcode-detector@2/dist/es/pure.min.js";



async function scan() {
  console.log("Running barcode.js");

  console.log("Verzögere die Ausführung um 4 Sekunden")
  await new Promise(resolve => setTimeout(resolve, 1500));

  const video = document.getElementById("video");
  const codeInput = document.getElementById("code-input");
  const newForm = document.getElementById("new-form");

  if (video === null || codeInput === null || newForm === null) {
    console.error("HTML Elements not found");
    return;
  }

  function showVideo() {
    video.style.display = "block";
  }

  let stream;
  if (navigator.mediaDevices !== undefined) {
    try {
      stream = await navigator.mediaDevices.getUserMedia({ video: true })
      video.srcObject = stream;
      await video.play();
    } catch (error) {
      console.log("Fehler beim Zugriff auf die Kamera", error);
      return;
    }
  } else {
    console.log("Keine Kamera verfügbar");
    return;
  }

  const barcodeDetector = new BarcodeDetector();
  const supportedFormats = await BarcodeDetector.getSupportedFormats();
  console.log("Unterstützte Barcode-Formate", supportedFormats);


  showVideo();

  let result = null;

  async function detectBarcodes() {
    try {
      const barcodes = await barcodeDetector.detect(video);
      if (barcodes.length === 1) {
        console.log("Gefundene Barcodes", barcodes);
        const barcode = barcodes[0]
        if (barcode.format === "ean_13") {
          result = barcode.rawValue;
          codeInput.value = barcode.rawValue
          console.log("Barcode", barcode.rawValue, "Format", barcode.format)
          if (stream) {
            console.log("Aufnahme wird angehalten")
            stream.getTracks().forEach(track => track.stop());
          }
        }
      }
    } catch (error) {
      console.error("Fehler bei der Barcodeerkennung", error)
    }
    finally {
      if (result === null) {
        requestAnimationFrame(detectBarcodes)
      } else {
        console.log("Frame Animation wird nicht fortgesetzt")
      }
    }
  };

  detectBarcodes();
}

scan();
