import React, { useState } from 'react';

function ReportViewer() {
  const [report, setReport] = useState(null);

  const handleFileUpload = (e) => {
    const file = e.target.files[0];
    const reader = new FileReader();
    
    reader.onload = (e) => {
      try {
        const data = JSON.parse(e.target.result);
        setReport(data);
      } catch (err) {
        console.error("Invalid report file:", err);
      }
    };
    
    reader.readAsText(file);
  };

  return (
    <div>
      <input type="file" accept=".json" onChange={handleFileUpload} />
      {report && (
        <div>
          <h2>Test Report: {report.meta.test_id}</h2>
          {/* Render metrics using charting library */}
        </div>
      )}
    </div>
  );
}

export default ReportViewer; 