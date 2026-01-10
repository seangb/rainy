document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('yearcomp-chart');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('yearcomp-json');
if (!pre) {
  console.error('Pre element with id "yearcomp-json" not found');
  return;
}
let yearCompTotals;
try {
  const jsonText = pre.textContent.trim();
  if (!jsonText || jsonText.startsWith('{{')) {
    console.error('Year comparison totals JSON is missing or not rendered.');
    return;
  }
  yearCompTotals = JSON.parse(jsonText);
} catch (e) {
  console.error('Failed to parse year comparison totals JSON:', e);
  return;
}
console.log(yearCompTotals);

const labels = yearCompTotals.map(item => item.Period);
// Assume labels is an array of periods like ['2025', '2025', ...]
// Compute the last period only
const now = new Date();
const lastPeriod = `${now.getFullYear()}`;

// For background color, highlight bars in the last period (including current)
const backgroundColors = labels.map(label =>
    lastPeriod === label ? 'red' : '#3a6ea5'
);

const data = {
    Total: yearCompTotals.map(item => item.TotalMM),
};
const ctx = canvas.getContext('2d');
if (!ctx) {
  console.error('Canvas context not found');
  return;
}
else {
  console.log('Canvas context successfully retrieved');
}

function renderChart(filteredTotals) {
    const filteredLabels = filteredTotals.map(item => item.Period);
    const filteredData = filteredTotals.map(item => item.TotalMM);
    const filteredBackgroundColors = filteredLabels.map(label =>
        lastPeriod === label ? 'red' : '#3a6ea5'
    );

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
renderChart(yearCompTotals);

});
// Ensure the script runs after the DOM is fully loaded
// This is already handled by the DOMContentLoaded event listener above