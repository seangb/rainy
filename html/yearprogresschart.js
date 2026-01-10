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

// Store the original data for filtering
const originalData = [...yearProgressTotals];

// Create the buttons container if it doesn't exist
let buttonsContainer = document.getElementById('chart-buttons');
if (!buttonsContainer) {
  buttonsContainer = document.createElement('div');
  buttonsContainer.id = 'chart-buttons';
  buttonsContainer.style.marginBottom = '10px';
  buttonsContainer.style.textAlign = 'center';
  canvas.parentNode.insertBefore(buttonsContainer, canvas);
}

// Add the filter button
const filterButton = document.createElement('button');
filterButton.textContent = 'Show Last 10 Years';
filterButton.style.marginRight = '10px';
filterButton.style.padding = '5px 10px';
filterButton.style.cursor = 'pointer';

// Add the reset button
const resetButton = document.createElement('button');
resetButton.textContent = 'Show All Years';
resetButton.style.padding = '5px 10px';
resetButton.style.cursor = 'pointer';

// Add buttons to container
buttonsContainer.appendChild(filterButton);
buttonsContainer.appendChild(resetButton);

// Function to create/update the chart
function createChart(filteredData) {
  // Create the line chart displaying yearProgressTotals.
  const allPeriods = new Set();
  filteredData.forEach(yearData => {
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

  let datasets = filteredData.map((yearData, idx) => ({
    label: yearData.Year,
    data: yearData.Totals.map(item => item.TotalMM),
    borderColor: getColor(idx, yearData.Year === currentYear),
    borderWidth: yearData.Year === currentYear ? 4 : 3, // Double thickness for current year
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
  }

  // Destroy existing chart if it exists
  if (window.yearProgressChart) {
    window.yearProgressChart.destroy();
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
}

// Initial chart creation
createChart(yearProgressTotals);

// Add filter functionality
filterButton.addEventListener('click', function() {
  const currentYear = new Date().getFullYear();
  const lastTenYears = currentYear - 9; // 10 years inclusive
  const filteredData = originalData.filter(yearData => yearData.Year >= lastTenYears);
  
  createChart(filteredData);
  filterButton.disabled = true;
  resetButton.disabled = false;
});

// Add reset functionality
resetButton.addEventListener('click', function() {
  createChart(originalData);
  filterButton.disabled = false;
  resetButton.disabled = true;
});

// Initially disable the reset button
resetButton.disabled = true;

});