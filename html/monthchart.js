document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('month-chart');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('monthly-json');
if (!pre) {
  console.error('Pre element with id "monthly-json" not found');
  return;
}

let monthlyTotals;
try {
  const jsonText = pre.textContent.trim();
  if (!jsonText || jsonText.startsWith('{{')) {
    console.error('Monthly totals JSON is missing or not rendered.');
    return;
  }
  monthlyTotals = JSON.parse(jsonText);
} catch (e) {
  console.error('Failed to parse monthly totals JSON:', e);
  return;
}

console.log(monthlyTotals);

const labels = monthlyTotals.map(item => item.Period);
const data = {
  Average: monthlyTotals.map(item => item.Average),
  LastTotal: monthlyTotals.map(item => item.LastTotal)
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
