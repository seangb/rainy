document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('quarter-comp');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('quarter-comp-json');
if (!pre) {
  console.error('Pre element with id "quarter-comp-json" not found');
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
// Assume labels is an array of periods like ['2025-Q1', '2025-Q2', '2025-Q3', ...]
// Compute the last 4 periods including the current quarter
const now = new Date();
const periodsSet = new Set();
for (let i = 0; i < 4; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i * 3, 1);
    const period = `${d.getFullYear()}-Q${Math.floor(d.getMonth() / 3) + 1}`;
    periodsSet.add(period);
}
const last4Periods = periodsSet;

// For background color, highlight bars in the last 4 quarters (including current)
const backgroundColors = labels.map(label =>
    last4Periods.has(label) ? 'red' : '#3a6ea5'
);

const data = {
    Total: quarterlyTotals.map(item => item.TotalMM),
};
const ctx = canvas.getContext('2d');
if (!ctx) {
  console.error('Canvas context not found');
  return;
}
else {
  console.log('Canvas context successfully retrieved');
}

const selector = document.createElement('select');
selector.id = 'quarter-filter';
const allOption = document.createElement('option');
allOption.value = '';
allOption.textContent = 'All Quarters';
selector.appendChild(allOption);

// Add Q1-Q4 options to selector
['Q1', 'Q2', 'Q3', 'Q4'].forEach(quarter => {
    const option = document.createElement('option');
    option.value = quarter;
    option.textContent = quarter;
    selector.appendChild(option);
});

// Insert selector before the canvas
canvas.parentNode.insertBefore(selector, canvas);

// Chart rendering function
function renderChart(filteredTotals) {
    const filteredLabels = filteredTotals.map(item => item.Period);
    const filteredData = filteredTotals.map(item => item.TotalMM);
    
    // Get the current quarter
    const now = new Date();
    const currentQuarter = `${now.getFullYear()}-Q${Math.floor(now.getMonth() / 3) + 1}`;
    
    const filteredBackgroundColors = filteredLabels.map(label => {
      if (label === currentQuarter) {
        return '#FFD600'; // Current quarter in dark yellow
      } else if (last4Periods.has(label) && label !== currentQuarter) {
        return 'red'; // Recent quarters (excluding current) in red
      } else {
        return '#3a6ea5'; // Other quarters in blue
      }
    });

    if (window.quarterChart) {
        window.quarterChart.destroy();
    }
    window.quarterChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: filteredLabels,
            datasets: [
                {
                    label: 'Total',
                    data: filteredData,
                    backgroundColor: filteredBackgroundColors,
                    borderWidth: 1
                }
            ]
        },
        options: {
            responsive: true,
            plugins: {
                legend: { position: 'top' },
                title: { display: true, text: 'Quarterly Totals Comparison' }
            }
        }
    });
}

// Initial render
renderChart(quarterlyTotals);

// Filter on selector change
selector.addEventListener('change', function () {
    const selectedQuarter = selector.value;
    if (!selectedQuarter) {
        renderChart(quarterlyTotals);
    } else {
        const filtered = quarterlyTotals.filter(item => item.Period.slice(5, 7) === selectedQuarter);
        renderChart(filtered);
    }
});
}
);
// Ensure the script runs after the DOM is fully loaded
// This is already handled by the DOMContentLoaded event listener above