document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('searchForm');
    const loading = document.getElementById('loading');
    const results = document.getElementById('results');
    const error = document.getElementById('error');
    const resultsBody = document.getElementById('resultsBody');
    const resultsInfo = document.getElementById('resultsInfo');
    const errorMessage = document.getElementById('errorMessage');
    const dateRangeInput = document.getElementById('dateRange');

    let selectedDates = [];

    flatpickr(dateRangeInput, {
        mode: "range",
        dateFormat: "M j",
        defaultDate: ["2025-01-15", "2025-01-20"],
        onReady: function(selectedDates, dateStr, instance) {
            const yearElement = instance.currentYearElement;
            if (yearElement) {
                yearElement.style.display = 'none';
            }
            const yearInput = instance.yearElements && instance.yearElements[0];
            if (yearInput) {
                yearInput.style.display = 'none';
            }
        },
        onOpen: function(selectedDates, dateStr, instance) {
            instance.jumpToDate(new Date(2025, 0, 15));
        },
        onMonthChange: function(selectedDates, dateStr, instance) {
            const yearElement = instance.currentYearElement;
            if (yearElement) {
                yearElement.style.display = 'none';
            }
        },
        onClose: function(selectedDatesArray) {
            selectedDates = selectedDatesArray;
        }
    });

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        if (selectedDates.length === 0) {
            showError('Please select at least one date');
            return;
        }

        const formData = new FormData(form);
        const params = new URLSearchParams();

        const startDate = selectedDates[0];
        const endDate = selectedDates.length > 1 ? selectedDates[1] : startDate;

        params.append('start_date', formatDate(startDate));
        params.append('end_date', formatDate(endDate));
        params.append('prefecture', formData.get('prefecture'));
        params.append('limit', formData.get('limit'));

        loading.style.display = 'block';
        results.style.display = 'none';
        error.style.display = 'none';

        try {
            const response = await fetch(`/api/search?${params.toString()}`);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            loading.style.display = 'none';

            if (data.length === 0) {
                showError('No results found for the selected criteria.');
                return;
            }

            displayResults(data, {
                startDate,
                endDate,
                prefecture: formData.get('prefecture')
            });

        } catch (err) {
            loading.style.display = 'none';
            showError(err.message);
        }
    });

    function formatDate(date) {
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        return `${month}-${day}`;
    }

    function displayResults(data, searchParams) {
        const prefectureText = searchParams.prefecture === 'all' ? 'All of Japan' :
            searchParams.prefecture.charAt(0).toUpperCase() + searchParams.prefecture.slice(1);

        const formatDisplayDate = (date) => {
            return date.toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric'
            });
        };

        const isSameDay = searchParams.startDate.toDateString() === searchParams.endDate.toDateString();
        const dateRangeText = isSameDay
            ? formatDisplayDate(searchParams.startDate)
            : `${formatDisplayDate(searchParams.startDate)} - ${formatDisplayDate(searchParams.endDate)}`;

        const dayCount = Math.ceil((searchParams.endDate - searchParams.startDate) / (1000 * 60 * 60 * 24)) + 1;

        resultsInfo.innerHTML = '';

        resultsBody.innerHTML = '';

        data.forEach(resort => {
            const row = document.createElement('tr');
            const rankClass = resort.Rank <= 3 ? ` class="rank-${resort.Rank}"` : '';

            const formatValue = (val) => val != null ? val : '-';
            const formatSnowfall = (val) => val != null ? `${val} cm` : '-';

            row.innerHTML = `
                <td${rankClass}>${resort.Rank}</td>
                <td>${resort.Name}</td>
                <td>${resort.Prefecture.charAt(0).toUpperCase() + resort.Prefecture.slice(1)}</td>
                <td><strong>${formatSnowfall(resort.AvgSnowfall)}</strong></td>
                <td>${formatValue(resort.YearsWithData)}</td>
                <td>${formatValue(resort.TopElevation)}</td>
                <td>${formatValue(resort.BaseElevation)}</td>
                <td>${formatValue(resort.VerticalDrop)}</td>
                <td>${formatValue(resort.NumCourses)}</td>
            `;

            row.dataset.rank = resort.Rank;
            row.dataset.name = resort.Name.toLowerCase();
            row.dataset.prefecture = resort.Prefecture.toLowerCase();
            row.dataset.snowfall = resort.AvgSnowfall != null ? resort.AvgSnowfall : -1;
            row.dataset.years = resort.YearsWithData != null ? resort.YearsWithData : -1;
            row.dataset.top = resort.TopElevation != null ? resort.TopElevation : -1;
            row.dataset.base = resort.BaseElevation != null ? resort.BaseElevation : -1;
            row.dataset.vertical = resort.VerticalDrop != null ? resort.VerticalDrop : -1;
            row.dataset.courses = resort.NumCourses != null ? resort.NumCourses : -1;

            resultsBody.appendChild(row);
        });

        results.style.display = 'block';
        results.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }

    function showError(message) {
        errorMessage.textContent = message;
        error.style.display = 'block';
    }

    let currentSort = { column: null, ascending: true };

    document.querySelectorAll('.sortable').forEach(header => {
        header.addEventListener('click', function() {
            const column = this.dataset.column;
            const isAscending = currentSort.column === column ? !currentSort.ascending : true;

            currentSort = { column, ascending: isAscending };

            sortTable(column, isAscending);
            updateSortArrows(this, isAscending);
        });
    });

    function sortTable(column, ascending) {
        const tbody = resultsBody;
        const rows = Array.from(tbody.querySelectorAll('tr'));

        rows.sort((a, b) => {
            let aVal = a.dataset[column];
            let bVal = b.dataset[column];

            if (column === 'name' || column === 'prefecture') {
                return ascending
                    ? aVal.localeCompare(bVal)
                    : bVal.localeCompare(aVal);
            } else {
                aVal = parseFloat(aVal);
                bVal = parseFloat(bVal);
                return ascending ? aVal - bVal : bVal - aVal;
            }
        });

        tbody.innerHTML = '';
        rows.forEach(row => tbody.appendChild(row));
    }

    function updateSortArrows(activeHeader, ascending) {
        document.querySelectorAll('.sortable .sort-arrow').forEach(arrow => {
            arrow.textContent = '';
            arrow.parentElement.classList.remove('sort-asc', 'sort-desc');
        });

        const arrow = activeHeader.querySelector('.sort-arrow');
        arrow.textContent = ascending ? ' ▲' : ' ▼';
        activeHeader.classList.add(ascending ? 'sort-asc' : 'sort-desc');
    }

    // ===== Peak Info Functionality =====
    const peakInfoForm = document.getElementById('peakInfoForm');
    const resortSelect = document.getElementById('resortSelect');
    const peakResults = document.getElementById('peakResults');
    const peakResultsBody = document.getElementById('peakResultsBody');
    const peakResultsTitle = document.getElementById('peakResultsTitle');

    // Load resort list on page load
    async function loadResortsList() {
        try {
            const response = await fetch('/api/resorts-with-peaks');
            if (!response.ok) {
                throw new Error('Failed to load resorts list');
            }
            const resorts = await response.json();

            // Sort resorts by name
            resorts.sort((a, b) => a.name.localeCompare(b.name));

            // Add options to select
            resorts.forEach(resort => {
                const option = document.createElement('option');
                option.value = resort.id;
                option.textContent = resort.name;
                resortSelect.appendChild(option);
            });
        } catch (err) {
            console.error('Error loading resorts:', err);
        }
    }

    // Load resorts on page load
    loadResortsList();

    // Handle peak info form submission
    peakInfoForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const resortId = resortSelect.value;

        loading.style.display = 'block';
        peakResults.style.display = 'none';
        results.style.display = 'none';
        error.style.display = 'none';

        try {
            const response = await fetch(`/api/peak-info?resort_id=${encodeURIComponent(resortId)}`);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            loading.style.display = 'none';

            if (resortId === 'all') {
                // Display all resorts with peaks
                displayPeakResults(data, true);
            } else {
                // Display single resort
                displayPeakResults([data], false);
            }

        } catch (err) {
            loading.style.display = 'none';
            showError(err.message);
        }
    });

    function displayPeakResults(data, isAllResorts) {
        peakResultsBody.innerHTML = '';

        if (isAllResorts) {
            peakResultsTitle.textContent = `Peak Snowfall Periods - All Resorts (${data.length})`;
        } else {
            peakResultsTitle.textContent = `Peak Snowfall Periods - ${data[0].Resort.name}`;
        }

        data.forEach(resortData => {
            const resort = resortData.Resort;
            const peaks = resortData.Peaks || [];

            if (peaks.length === 0) return;

            const row = document.createElement('tr');

            const formatValue = (val) => val != null ? val : '-';

            // Create peaks cell content
            let peaksHtml = '';
            peaks.forEach(peak => {
                const dateRange = peak.start_date === peak.end_date
                    ? peak.start_date
                    : `${peak.start_date} - ${peak.end_date}`;

                peaksHtml += `
                    <div class="peak-entry">
                        <span class="peak-rank">#${peak.peak_rank}</span>
                        <span class="peak-dates">${dateRange}</span>
                        <span class="peak-snowfall">${Math.round(peak.total_period_snowfall)}cm (${peak.avg_daily_snowfall.toFixed(1)}cm/day)</span>
                        <span class="peak-confidence confidence-${peak.confidence_level}">${peak.confidence_level}</span>
                    </div>
                `;
            });

            const topPeakSnowfall = peaks.length > 0 ? peaks[0].total_period_snowfall : 0;

            row.innerHTML = `
                <td><div class="resort-name">${resort.name}</div></td>
                <td>${resort.prefecture}</td>
                <td>${formatValue(resort.top_elevation_m)}</td>
                <td>${formatValue(resort.base_elevation_m)}</td>
                <td>${formatValue(resort.vertical_m)}</td>
                <td>${formatValue(resort.num_courses)}</td>
                <td class="peaks-cell" data-peak-snowfall="${topPeakSnowfall}">${peaksHtml}</td>
            `;

            // Add data attributes for sorting
            row.dataset.name = resort.name.toLowerCase();
            row.dataset.prefecture = resort.prefecture.toLowerCase();
            row.dataset.top = resort.top_elevation_m != null ? resort.top_elevation_m : -1;
            row.dataset.base = resort.base_elevation_m != null ? resort.base_elevation_m : -1;
            row.dataset.vertical = resort.vertical_m != null ? resort.vertical_m : -1;
            row.dataset.courses = resort.num_courses != null ? resort.num_courses : -1;
            row.dataset.peaks = topPeakSnowfall;

            peakResultsBody.appendChild(row);
        });

        peakResults.style.display = 'block';
        peakResults.scrollIntoView({ behavior: 'smooth', block: 'start' });

        // Setup sorting for peak results table
        setupPeakTableSorting();
    }

    function setupPeakTableSorting() {
        const peakTable = document.getElementById('peakResultsTable');
        const headers = peakTable.querySelectorAll('.sortable');

        headers.forEach(header => {
            // Remove any existing listeners
            const newHeader = header.cloneNode(true);
            header.parentNode.replaceChild(newHeader, header);
        });

        peakTable.querySelectorAll('.sortable').forEach(header => {
            header.addEventListener('click', function() {
                const column = this.dataset.column;
                const tbody = peakResultsBody;
                const rows = Array.from(tbody.querySelectorAll('tr'));
                const isAscending = this.classList.contains('asc');

                // Clear all sort indicators
                peakTable.querySelectorAll('.sortable').forEach(h => {
                    h.classList.remove('asc', 'desc');
                });

                // Set current sort indicator
                this.classList.add(isAscending ? 'desc' : 'asc');

                rows.sort((a, b) => {
                    let aVal = a.dataset[column];
                    let bVal = b.dataset[column];

                    if (column === 'name' || column === 'prefecture') {
                        return isAscending
                            ? bVal.localeCompare(aVal)
                            : aVal.localeCompare(bVal);
                    } else {
                        aVal = parseFloat(aVal);
                        bVal = parseFloat(bVal);
                        return isAscending ? bVal - aVal : aVal - bVal;
                    }
                });

                tbody.innerHTML = '';
                rows.forEach(row => tbody.appendChild(row));
            });
        });
    }
});
