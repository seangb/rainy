document.addEventListener("DOMContentLoaded", function(event) {
// Ensure Chart.js is loaded
if (typeof Chart === 'undefined') {
  console.error('Chart.js is not loaded. Please ensure the library is included in your HTML.');
  return;
}

// Ensure the canvas element exists
const canvas = document.getElementById('yearprogress-chart');
if (!canvas) {
  console.error('Canvas element not found');
  return;
}

// Parse the JSON data from the pre element
const pre = document.getElementById('yearprogress-json');
if (!pre) {
  console.error('Pre element with id "yearprogress-json" not found');
  return;
}
let yearProgressTotals;
try {
  const jsonText = pre.textContent.trim();
  if (!jsonText || jsonText.startsWith('{{')) {
    console.error('Year progress totals JSON is missing or not rendered.');
    return;
  }
  yearProgressTotals = JSON.parse(jsonText);
} catch (e) {
  console.error('Failed to parse year progress totals JSON:', e);
  return;
}
console.log(yearProgressTotals);

// Create the line chart displaying yearProgressTotals.
// Make the line for the current year red, and make it so that it is visible on top of the other lines
// Remove the circles for the individual data points
// Make each year's line a distinct color. Make the current year's line black.
const allPeriods = new Set();
yearProgressTotals.forEach(yearData => {
  yearData.Totals.forEach(item => allPeriods.add(item.Period));
});
const labels = Array.from(allPeriods).sort();

// Determine the current year
const currentYear = new Date().getFullYear();

// Generate a distinct color for each year, and make the current year's line double the thickness
function getColor(index, isCurrent) {
  if (isCurrent) return 'rgba(0, 0, 0, 1)'; // Black for current year
  // Generate HSL colors for distinction
  const hue = (index * 60) % 360;
  return `hsl(${hue}, 70%, 50%)`;
}

let datasets = yearProgressTotals.map((yearData, idx) => ({
  label: yearData.Year,
  data: yearData.Totals.map(item => item.TotalMM),
  borderColor: getColor(idx, yearData.Year === currentYear),
  borderWidth: yearData.Year === currentYear ? 4 : 3, // Double thickness for current year
  // backgroundColor: 'rgba(237, 240, 243, 0.2)',
  backgroundColor: 'rgba(255, 255, 255, 0.0)',
  fill: true,
  tension: 0.4, // Smooth the line
  pointRadius: 0, // Remove circles for data points
}));

// Move the current year's dataset to the end of the array so it appears on top
datasets = datasets.sort((a, b) => (a.label === currentYear ? 1 : b.label === currentYear ? -1 : 0));

const data = {
  labels: labels,
  datasets: datasets,
};

const ctx = canvas.getContext('2d');
if (!ctx) {
  console.error('Canvas context not found');
  return;
} else {
  console.log('Canvas context successfully retrieved');
}

window.yearProgressChart = new Chart(ctx, {
  type: 'line',
  data: data,
  options: {
    responsive: true,
    plugins: {
      title: {
        display: true,
        text: 'Year Progress Chart'
      }
    },
    scales: {
      y: {
        beginAtZero: true
      }
    }
  }
});

});