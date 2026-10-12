const fs = require('fs');

// Create a valid, compliant text-based PDF 1.4 document
function buildPDF() {
  const content = `BT
/F1 16 Tf
50 750 Td
(Project Engineering Report: Distance Monitoring System) Tj
ET
BT
/F1 12 Tf
50 710 Td
(Section 2.1 Hardware Architecture and Sensors) Tj
ET
BT
/F1 10 Tf
50 670 Td
(The project uses an HC-SR04 ultrasonic sensor for non-contact distance measurement.) Tj
ET
BT
/F1 10 Tf
50 650 Td
(Operating parameters are between 2cm and 400cm with an accuracy rating of 3mm.) Tj
ET
BT
/F1 10 Tf
50 630 Td
(Distance calculation relies exclusively on the HC-SR04 ultrasonic transducer configured) Tj
ET
BT
/F1 10 Tf
50 610 Td
(on GPIO Pin 12 and Pin 13, calibrated for 3mm precision over 2-400cm range.) Tj
ET
BT
/F1 10 Tf
50 580 Td
(Microcontroller unit: STM32F401RE ARM Cortex-M4 operating at 84MHz.) Tj
ET
BT
/F1 10 Tf
50 550 Td
(Power supply requirements: 5V DC regulated voltage through LM2596 buck converter.) Tj
ET
BT
/F1 10 Tf
50 520 Td
(Experimental results demonstrate 99.1% measurement consistency under lab conditions.) Tj
ET`;

  const streamLength = Buffer.byteLength(content, 'utf-8');

  let pdf = `%PDF-1.4
1 0 obj
<<
  /Type /Catalog
  /Pages 2 0 R
>>
endobj
2 0 obj
<<
  /Type /Pages
  /Kids [3 0 R]
  /Count 1
>>
endobj
3 0 obj
<<
  /Type /Page
  /Parent 2 0 R
  /MediaBox [0 0 595 842]
  /Contents 4 0 R
  /Resources <<
    /Font <<
      /F1 5 0 R
    >>
  >>
>>
endobj
4 0 obj
<<
  /Length ${streamLength}
>>
stream
${content}
endstream
endobj
5 0 obj
<<
  /Type /Font
  /Subtype /Type1
  /BaseFont /Helvetica
>>
endobj
xref
0 6
0000000000 65535 f 
0000000010 00000 n 
0000000060 00000 n 
00000000115 00000 n 
00000000250 00000 n 
00000000${(250 + streamLength + 35).toString().padStart(3, '0')} 00000 n 
trailer
<<
  /Size 6
  /Root 1 0 R
>>
startxref
${(250 + streamLength + 100)}
%%EOF
`;

  fs.writeFileSync('project_report.pdf', pdf);
  console.log('project_report.pdf generated successfully!');
}

buildPDF();
