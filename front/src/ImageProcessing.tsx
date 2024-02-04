// ImageProcessing.tsx

import React, { useState, ChangeEvent } from 'react';
import axios from 'axios';

export interface ImageProcessingProps {
  // Propsの定義（必要に応じて追加）
}

const ImageProcessing: React.FC<ImageProcessingProps> = (props) => {
  const [image, setImage] = useState<File | null>(null);
  const [resize, setResize] = useState<boolean>(false);
  const [antialiasing, setAntialiasing] = useState<boolean>(false);
  const [smoothing, setSmoothing] = useState<boolean>(false);
  const [gaussian, setGaussian] = useState<boolean>(false);
  const [unsharpMask, setUnsharpMask] = useState<boolean>(false);
  const [processedImage, setProcessedImage] = useState<string | null>("http://localhost:8080/processed_image.png");

  const handleImageChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const selectedImage = e.target.files[0];
      setImage(selectedImage);
      setProcessedImage(null); // Reset processed image when a new image is selected
    }
  };

  const handleProcessing = async () => {
    if (!image) {
      alert('Please select an image.');
      return;
    }

    const formData = new FormData();
    formData.append('image', image);
    formData.append('resize', resize.toString());
    formData.append('antialiasing', antialiasing.toString());
    formData.append('smoothing', smoothing.toString());
    formData.append('gaussian', gaussian.toString());
    formData.append('unsharpMask', unsharpMask.toString());

    try {
      const response = await axios.post('http://localhost:8080/process', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        responseType: 'arraybuffer', // Specify response type as arraybuffer
      });
    } catch (error) {
      console.error('Error processing image:', error);
    }
    setProcessedImage(null);

  };

  const handleUpdateImage = async () => {
    try {
      // Fetch the processed image again
      const response = await axios.get("http://localhost:8080/processed_image.png", {
        responseType: 'arraybuffer', // Specify response type as arraybuffer
      });

      // Convert the arraybuffer to base64 to display as image
      const base64 = btoa(new Uint8Array(response.data).reduce((data, byte) => data + String.fromCharCode(byte), ''));
      setProcessedImage(`data:image/png;base64,${base64}`);
    } catch (error) {
      console.error('Error updating image:', error);
    }
  };

  return (
    <div>
      <input type="file" accept="image/*" onChange={handleImageChange} />
      <label>
        Resize
        <input type="checkbox" checked={resize} onChange={() => setResize(!resize)} />
      </label>
      <label>
        Antialiasing
        <input type="checkbox" checked={antialiasing} onChange={() => setAntialiasing(!antialiasing)} />
      </label>
      <label>
        Smoothing
        <input type="checkbox" checked={smoothing} onChange={() => setSmoothing(!smoothing)} />
      </label>
      <label>
        Gaussian
        <input type="checkbox" checked={gaussian} onChange={() => setGaussian(!gaussian)} />
      </label>
      <label>
        Unsharp Mask
        <input type="checkbox" checked={unsharpMask} onChange={() => setUnsharpMask(!unsharpMask)} />
      </label>
      <button onClick={handleProcessing}>Process Image</button>
      <button onClick={handleUpdateImage}>Update Image</button>

      {processedImage && (
        <div>
          <h2>Processed Image Preview</h2>
          <img src={processedImage} alt="Processed" style={{ maxWidth: '100%', maxHeight: '400px' }} />
        </div>
      )}
    </div>
  );
};

export default ImageProcessing;
