import React from 'react';
import { useParams } from 'react-router-dom';

const DocumentEditor: React.FC = () => {
  const { fileId } = useParams<{ fileId: string }>();
  const token = localStorage.getItem('token') || '';

  const src = fileId
    ? `/api/files/${fileId}/edit?token=${encodeURIComponent(token)}`
    : '';

  return (
    <div style={{ width: '100%', height: '100vh', margin: 0, padding: 0 }}>
      {src ? (
        <iframe
          src={src}
          title="OnlyOffice Editor"
          style={{ width: '100%', height: '100%', border: 'none' }}
          allow="clipboard-read; clipboard-write"
        />
      ) : (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
          Invalid file ID
        </div>
      )}
    </div>
  );
};

export default DocumentEditor;