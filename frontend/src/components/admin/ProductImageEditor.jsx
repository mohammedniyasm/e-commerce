import { useEffect, useRef, useState } from 'react';
import { Button } from '../common/Button';
import './ProductImageEditor.css';

const CROP_WIDTH = 360;
const CROP_HEIGHT = 480;
const MAX_OUTPUT_EDGE = 1600;
const MAX_UPLOAD_BYTES = 5 * 1024 * 1024;
const SUPPORTED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp'];

function getCrop(image, zoom, panX, panY) {
  const scale = Math.max(
    CROP_WIDTH / image.naturalWidth,
    CROP_HEIGHT / image.naturalHeight,
  ) * zoom;
  const width = image.naturalWidth * scale;
  const height = image.naturalHeight * scale;
  const left = (CROP_WIDTH - width) / 2 + panX;
  const top = (CROP_HEIGHT - height) / 2 + panY;
  const sourceWidth = CROP_WIDTH / scale;
  const sourceHeight = CROP_HEIGHT / scale;

  return {
    left,
    top,
    width,
    height,
    sourceX: Math.max(0, Math.min(image.naturalWidth - sourceWidth, -left / scale)),
    sourceY: Math.max(0, Math.min(image.naturalHeight - sourceHeight, -top / scale)),
    sourceWidth,
    sourceHeight,
  };
}

export function ProductImageEditor({ file, src, onCancel, onApply }) {
  const canvasRef = useRef(null);
  const imageRef = useRef(null);
  const dragStartRef = useRef(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    let active = true;
    const image = new Image();
    image.onload = () => {
      if (!active) return;
      imageRef.current = image;
      setIsLoaded(true);
    };
    image.onerror = () => {
      if (active) setError('This image could not be opened for cropping.');
    };
    image.src = src;

    return () => {
      active = false;
      imageRef.current = null;
    };
  }, [src]);

  useEffect(() => {
    const image = imageRef.current;
    const context = canvasRef.current?.getContext('2d');
    if (!image || !context) return;

    const crop = getCrop(image, zoom, pan.x, pan.y);
    context.clearRect(0, 0, CROP_WIDTH, CROP_HEIGHT);
    context.drawImage(image, crop.left, crop.top, crop.width, crop.height);
  }, [isLoaded, zoom, pan]);

  function handlePointerDown(event) {
    dragStartRef.current = {
      x: event.clientX,
      y: event.clientY,
      panX: pan.x,
      panY: pan.y,
    };
    event.currentTarget.setPointerCapture(event.pointerId);
  }

  function handlePointerMove(event) {
    const start = dragStartRef.current;
    const image = imageRef.current;
    if (!start || !image) return;

    const rect = event.currentTarget.getBoundingClientRect();
    const crop = getCrop(image, zoom, start.panX, start.panY);
    const maxX = Math.max(0, (crop.width - CROP_WIDTH) / 2);
    const maxY = Math.max(0, (crop.height - CROP_HEIGHT) / 2);
    const nextX = start.panX + ((event.clientX - start.x) / rect.width) * CROP_WIDTH;
    const nextY = start.panY + ((event.clientY - start.y) / rect.height) * CROP_HEIGHT;
    setPan({
      x: Math.max(-maxX, Math.min(maxX, nextX)),
      y: Math.max(-maxY, Math.min(maxY, nextY)),
    });
  }

  function getOutputSize() {
    const image = imageRef.current;
    if (!image) return null;
    const crop = getCrop(image, zoom, pan.x, pan.y);
    const scale = Math.min(
      1,
      MAX_OUTPUT_EDGE / Math.max(crop.sourceWidth, crop.sourceHeight),
    );
    return {
      width: Math.max(1, Math.round(crop.sourceWidth * scale)),
      height: Math.max(1, Math.round(crop.sourceHeight * scale)),
    };
  }

  function handleApply() {
    const image = imageRef.current;
    if (!image) return;
    if (!SUPPORTED_IMAGE_TYPES.includes(file.type)) {
      setError('Only JPEG, PNG, and WebP images can be cropped.');
      return;
    }

    const crop = getCrop(image, zoom, pan.x, pan.y);
    const outputSize = getOutputSize();
    const output = document.createElement('canvas');
    output.width = outputSize.width;
    output.height = outputSize.height;
    const context = output.getContext('2d');
    if (!context) {
      setError('Image cropping is not supported by this browser.');
      return;
    }

    setIsSaving(true);
    setError('');
    context.drawImage(
      image,
      crop.sourceX,
      crop.sourceY,
      crop.sourceWidth,
      crop.sourceHeight,
      0,
      0,
      output.width,
      output.height,
    );

    output.toBlob(
      (blob) => {
        if (!blob) {
          setError('The cropped image could not be created.');
        } else if (blob.size > MAX_UPLOAD_BYTES) {
          setError('The cropped image is over 5 MB. Reduce the zoom and try again.');
        } else {
          onApply(new File([blob], file.name, { type: file.type, lastModified: Date.now() }));
        }
        setIsSaving(false);
      },
      file.type,
      0.92,
    );
  }

  const outputSize = getOutputSize();

  return (
    <section className="admin-image-editor" aria-label="Crop image for product listings">
      <div className="admin-image-editor__heading">
        <div>
          <strong>Product listing crop</strong>
          <p>3:4 portrait crop, matching the storefront product cards.</p>
        </div>
        <span>3:4</span>
      </div>
      <div
        className="admin-image-editor__canvas-wrap"
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={() => {
          dragStartRef.current = null;
        }}
        onPointerCancel={() => {
          dragStartRef.current = null;
        }}
      >
        <canvas
          ref={canvasRef}
          width={CROP_WIDTH}
          height={CROP_HEIGHT}
          role="img"
          aria-label={`3:4 crop preview for ${file.name}`}
        />
      </div>
      <label className="admin-image-editor__zoom">
        <span>Zoom {zoom.toFixed(1)}×</span>
        <input
          type="range"
          min="1"
          max="3"
          step="0.1"
          value={zoom}
          onChange={(event) => setZoom(Number(event.target.value))}
        />
      </label>
      {outputSize && (
        <p className="admin-image-editor__dimensions">
          Output: {outputSize.width} × {outputSize.height}px
          <span> · max 1600px, no upscaling</span>
        </p>
      )}
      {error && <p className="admin-image-editor__error" role="alert">{error}</p>}
      <div className="admin-image-editor__actions">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSaving}>
          Cancel
        </Button>
        <Button
          type="button"
          variant="primary"
          onClick={handleApply}
          disabled={!isLoaded || isSaving}
          isLoading={isSaving}
        >
          Apply Crop
        </Button>
      </div>
    </section>
  );
}
