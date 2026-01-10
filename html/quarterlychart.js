document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('quarterly-chart');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('quarterly-json');
if (!pre) {
  console.error('Pre element with id "quarterly-json" not found');
  return;
}

let quarterlyTotals;
try {
  const jsonText = pre.textContent.trim();
  if (!jsonText || jsonText.startsWith('{{')) {
    console.error('Quarterly totals JSON is missing or not rendered.');
    return;
  }
  quarterlyTotals = JSON.parse(jsonText);
} catch (e) {
  console.error('Failed to parse quarterly totals JSON:', e);
  return;
}

console.log(quarterlyTotals);

const labels = quarterlyTotals.map(item => item.Period);
const data = {
  Average: quarterlyTotals.map(item => item.Average),
  LastTotal: quarterlyTotals.map(item => item.LastTotal)
};

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
    datasets: [
      {
        label: 'Average',
        data: data.Average,
        borderWidth: 1
      },
      {
        label: 'Last Total',
        data: data.LastTotal,
        borderWidth: 1
      }
    ]
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
