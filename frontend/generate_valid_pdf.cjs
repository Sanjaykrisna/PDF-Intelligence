const fs = require('fs');
const { PDFDocument, StandardFonts, rgb } = require('pdf-lib');

async function createPDF() {
  const pdfDoc = await PDFDocument.create();
  const font = await pdfDoc.embedFont(StandardFonts.Helvetica);
  const fontBold = await pdfDoc.embedFont(StandardFonts.HelveticaBold);

  // Page 1: Title & Overview
  const page1 = pdfDoc.addPage([595, 842]);
  page1.drawText('Project Engineering Report: Distance Monitoring System', {
    x: 50,
    y: 780,
    size: 16,
    font: fontBold,
    color: rgb(0.1, 0.45, 0.91),
  });

  page1.drawText('Section 1. Executive Summary and Problem Statement', {
    x: 50,
    y: 740,
    size: 13,
    font: fontBold,
    color: rgb(0.13, 0.13, 0.14),
  });

  const p1Text = [
    'Traditional proximity detection systems often suffer from high latency and limited accuracy in',
    'dusty or variable lighting industrial environments. This project proposes an autonomous micro-',
    'controller system designed for precise, non-contact distance monitoring and obstacle avoidance.',
    'The core architecture integrates hardware sensing with embedded filtering algorithms.',
  ];
  let y = 715;
  for (const line of p1Text) {
    page1.drawText(line, { x: 50, y, size: 10, font, color: rgb(0.2, 0.2, 0.2) });
    y -= 18;
  }

  // Page 1: Section 2
  y -= 15;
  page1.drawText('Section 2.1 Hardware Architecture and Ultrasonic Sensor', {
    x: 50,
    y,
    size: 13,
    font: fontBold,
    color: rgb(0.13, 0.13, 0.14),
  });
  y -= 25;

  const p2Text = [
    'The project uses an HC-SR04 ultrasonic sensor for non-contact distance measurement, operating',
    'between 2cm and 400cm with an accuracy rating of 3mm.',
    'Distance calculation relies exclusively on the HC-SR04 ultrasonic transducer configured on GPIO',
    'Pin 12 (Trigger) and Pin 13 (Echo), calibrated for 3mm precision over the full 2cm to 400cm range.',
    'The microcontroller unit is an STM32F401RE ARM Cortex-M4 operating at 84MHz.',
    'Power is supplied via a 5V DC regulated line with an LM2596 buck regulator ensuring stable excitation.',
  ];
  for (const line of p2Text) {
    page1.drawText(line, { x: 50, y, size: 10, font, color: rgb(0.2, 0.2, 0.2) });
    y -= 18;
  }

  // Page 2: Methodology & Results
  const page2 = pdfDoc.addPage([595, 842]);
  page2.drawText('Section 3. Methodology, Algorithms, and Experimental Results', {
    x: 50,
    y: 780,
    size: 14,
    font: fontBold,
    color: rgb(0.13, 0.13, 0.14),
  });

  const p3Text = [
    'The acquisition methodology employs a median filtering window of size 5 to reject echo noise and',
    'multipath acoustic reflections. Timer capture registers measure the pulse high duration with',
    'sub-microsecond resolution.',
    '',
    'Results demonstrate 99.1% measurement consistency under lab test conditions across 500 samples.',
    'Advantages include low component cost, immunity to ambient light variations, and rapid response',
    'time under 60 milliseconds per measurement cycle.',
    '',
    'Future scope includes implementing dual-sensor array triangulation and wireless telemetry via BLE.',
  ];
  y = 740;
  for (const line of p3Text) {
    if (line) {
      page2.drawText(line, { x: 50, y, size: 10, font, color: rgb(0.2, 0.2, 0.2) });
    }
    y -= 18;
  }

  const pdfBytes = await pdfDoc.save();
  fs.writeFileSync('project_report.pdf', pdfBytes);
  console.log('Valid project_report.pdf created with 2 pages and text streams!');
}

createPDF();
