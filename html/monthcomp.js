document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('month-comp');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('month-comp-json');
if (!pre) {
  console.error('Pre element with id "month-comp-json" not found');
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
// Assume labels is an array of periods like ['2025-04', '2025-05', '2025-06', ...]
// Compute the last 12 periods including the current month
const now = new Date();
const periodsSet = new Set();
for (let i = 0; i < 12; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const period = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
    periodsSet.add(period);
}
const last12Periods = periodsSet;

// For background color, highlight bars in the last 12 periods (including current)
const backgroundColors = labels.map(label =>
    last12Periods.has(label) ? 'red' : '#3a6ea5'
);

const data = {
    Total: monthlyTotals.map(item => item.TotalMM),
};
const ctx = canvas.getContext('2d');
if (!ctx) {
  console.error('Canvas context not found');
  return;
}
else {
  console.log('Canvas context successfully retrieved');
}
// Create a selector for months
const uniqueMonths = [...new Set(monthlyTotals.map(item => {
    // Extract month as 'MM' from 'YYYY-MM'
    return item.Period.slice(5, 7);
}))];

// Map month numbers to short names
const monthNames = {
    '01': 'Jan', '02': 'Feb', '03': 'Mar', '04': 'Apr',
    '05': 'May', '06': 'Jun', '07': 'Jul', '08': 'Aug',
    '09': 'Sep', '10': 'Oct', '11': 'Nov', '12': 'Dec'
};

const selector = document.createElement('select');
selector.id = 'month-filter';
const allOption = document.createElement('option');
allOption.value = '';
allOption.textContent = 'All Months';
selector.appendChild(allOption);

// Sort uniqueMonths in calendar order before adding to selector
uniqueMonths
    .sort((a, b) => parseInt(a, 10) - parseInt(b, 10))
    .forEach(monthNum => {
        const option = document.createElement('option');
        option.value = monthNum;
        option.textContent = monthNames[monthNum] || monthNum;
        selector.appendChild(option);
    });

// Insert selector before the canvas
canvas.parentNode.insertBefore(selector, canvas);

// Chart rendering function
function renderChart(filteredTotals) {
    const filteredLabels = filteredTotals.map(item => item.Period);
    const filteredData = filteredTotals.map(item => item.TotalMM);
    const currentPeriod = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
    const filteredBackgroundColors = filteredLabels.map(label => {
        if (label === currentPeriod) return '#FFD600'; // dark yellow for current month
        if (last12Periods.has(label)) return '#DE3163';
        return '#3a6ea5';
    });

    if (window.monthChart) {
        window.monthChart.destroy();
    }
    window.monthChart = new Chart(ctx, {
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
                title: { display: true, text: 'Monthly Totals Comparison' }
            }
        }
    });
}

// Initial render
renderChart(monthlyTotals);

// Filter on selector change
selector.addEventListener('change', function () {
    const selectedMonth = selector.value;
    if (!selectedMonth) {
        renderChart(monthlyTotals);
    } else {
        const filtered = monthlyTotals.filter(item => item.Period.slice(5, 7) === selectedMonth);
        renderChart(filtered);
    }
});
}
);
// Ensure the script runs after the DOM is fully loaded
// This is already handled by the DOMContentLoaded event listener above