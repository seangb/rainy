document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('halfyearcomp-chart');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('halfyearcomp-json');
if (!pre) {
  console.error('Pre element with id "halfyearcomp-json" not found');
  return;
}
let halfYearTotals;
try {
  const jsonText = pre.textContent.trim();
  if (!jsonText || jsonText.startsWith('{{')) {
    console.error('Half-year totals JSON is missing or not rendered.');
    return;
  }
  halfYearTotals = JSON.parse(jsonText);
} catch (e) {
  console.error('Failed to parse half-year totals JSON:', e);
  return;
}
console.log(halfYearTotals);

const labels = halfYearTotals.map(item => item.Period);
// Assume labels is an array of periods like ['2025-H1', '2025-H2', ...]
// Compute the last 4 periods including the current half-year
const now = new Date();
const periodsSet = new Set();
for (let i = 0; i < 4; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i * 6, 1);
    const period = `${d.getFullYear()}-H${Math.floor(d.getMonth() / 6) + 1}`;
    periodsSet.add(period);
}
const last4Periods = periodsSet;

// For background color, highlight bars in the last 4 half-years (including current)
const backgroundColors = labels.map(label =>
    last4Periods.has(label) ? 'red' : '#3a6ea5'
);

const data = {
    Total: halfYearTotals.map(item => item.TotalMM),
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
selector.id = 'halfyear-filter';
const allOption = document.createElement('option');
allOption.value = '';
allOption.textContent = 'All Half-Years';
selector.appendChild(allOption);

// Add H1-H2 options to selector
['H1', 'H2'].forEach(halfYear => {
    const option = document.createElement('option');
    option.value = halfYear;
    option.textContent = halfYear;
    selector.appendChild(option);
});

// Insert selector before the canvas
canvas.parentNode.insertBefore(selector, canvas);

// Chart rendering function
function renderChart(filteredTotals) {
  const filteredLabels = filteredTotals.map(item => item.Period);
  const filteredData = filteredTotals.map(item => item.TotalMM);
  
  // Get current half-year
  const now = new Date();
  const currentHalfYear = `${now.getFullYear()}-H${Math.floor(now.getMonth() / 6) + 1}`;
  
  // Modify background colors - current half-year is dark yellow, recent ones are red
  const filteredBackgroundColors = filteredLabels.map(label => {
    if (label === currentHalfYear) {
      return '#FFD600'; // Dark yellow for current half-year
    } else if (last4Periods.has(label)) {
      return 'red'; // Red for other recent periods
    } else {
      return '#3a6ea5'; // Blue for older periods
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
        title: { display: true, text: 'Half-Year Totals Comparison' }
      }
    }
  });
}

// Initial render
renderChart(halfYearTotals);

// Filter on selector change
selector.addEventListener('change', function () {
    const selectedHalfyear = selector.value;
    if (!selectedHalfyear) {
        renderChart(halfYearTotals);
    } else {
        const filtered = halfYearTotals.filter(item => item.Period.slice(5, 7) === selectedHalfyear);
        renderChart(filtered);
    }
});
}
);
// Ensure the script runs after the DOM is fully loaded
// This is already handled by the DOMContentLoaded event listener above