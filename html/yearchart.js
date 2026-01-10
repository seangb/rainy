document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('myChart');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('yearly-json');
if (!pre) {
  console.error('Pre element with id "yearly-json" not found');
  return;
}

let yearlyTotals;
try {
  const jsonText = pre.textContent.trim();
  if (!jsonText || jsonText.startsWith('{{')) {
    console.error('Yearly totals JSON is missing or not rendered.');
    return;
  }
  yearlyTotals = JSON.parse(jsonText);
} catch (e) {
  console.error('Failed to parse yearly totals JSON:', e);
  return;
}

console.log(yearlyTotals);

const labels = yearlyTotals.map(item => item.Period);
const data = yearlyTotals.map(item => item.TotalMM);

const ctx = canvas.getContext('2d');
if (!ctx) {
  console.error('Canvas context not found');
  return;
} else {
  console.log('Canvas context successfully retrieved');
}

new Chart(ctx, {
  type: 'bar',
  data: {
    labels: labels,
    datasets: [{
      label: 'mm Totals',
      data: data,
      borderWidth: 1
    }]
  },
  options: {
    scales: {
      y: {
        beginAtZero: true
      }
    }
  }
});
});
