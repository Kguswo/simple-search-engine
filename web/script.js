console.log('=== script.js 로드 완료 ===');

const API_BASE_URL = 'http://localhost:8080';

const searchInput = document.getElementById('searchInput');
const searchBtn = document.getElementById('searchBtn');
const resultsDiv = document.getElementById('results');
const loadingDiv = document.getElementById('loading');
const searchInfoDiv = document.getElementById('searchInfo');

// 검색 버튼 클릭
searchBtn.addEventListener('click', performSearch);

// Enter 키 입력
searchInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        performSearch();
    }
});

async function performSearch() {
    console.log('=== 검색 시작 ===');
    const query = searchInput.value.trim();
    console.log('검색어:', query);
    
    if (!query) {
        alert('검색어를 입력해주세요!');
        return;
    }
    
    // 로딩 표시
    loadingDiv.style.display = 'block';
    resultsDiv.innerHTML = '';
    searchInfoDiv.style.display = 'none';

    try {
        // 스마트 검색 API 호출
        const response = await fetch(`${API_BASE_URL}/search/smart?q=${encodeURIComponent(query)}`);
        
        if (!response.ok) {
            throw new Error('검색 실패');
        }

        const data = await response.json();
        
        loadingDiv.style.display = 'none';
        
        // 검색 방법 표시
        if (data.search_method && data.corrected_query !== query) {
            searchInfoDiv.innerHTML = `
                🔍 <strong>${data.search_method}</strong>로 검색됨: 
                <span style="text-decoration: line-through;">${query}</span> 
                → <strong>${data.corrected_query}</strong>
            `;
            searchInfoDiv.style.display = 'block';
        } else if (data.search_method) {
            searchInfoDiv.innerHTML = `🔍 <strong>${data.search_method}</strong>`;
            searchInfoDiv.style.display = 'block';
        }
        
        displayResults(data.results, data.query);
        
    } catch (error) {
        loadingDiv.style.display = 'none';
        resultsDiv.innerHTML = `
            <div class="no-results">
                ❌ 검색 중 오류가 발생했습니다.<br>
                API 서버가 실행 중인지 확인해주세요.
            </div>
        `;
        console.error('Search error:', error);
    }
}

function displayResults(results, query) {
    if (!results || results.length === 0) {
        resultsDiv.innerHTML = `
            <div class="no-results">
                😢 "${query}"에 대한 검색 결과가 없습니다.
            </div>
        `;
        return;
    }

    resultsDiv.innerHTML = '';
    
    results.forEach(result => {
        const resultItem = document.createElement('div');
        resultItem.className = 'result-item';
        
        // 제목 (클릭 가능한 링크)
        const title = document.createElement('div');
        title.className = 'result-title';
        const link = document.createElement('a');
        link.href = result.url;
        link.target = '_blank';
        link.textContent = result.title;
        title.appendChild(link);
        resultItem.appendChild(title);
        
        // 내용 (150자 제한)
        const content = document.createElement('div');
        content.className = 'result-content';
        const contentText = result.content || '';
        content.textContent = contentText.length > 150 
            ? contentText.substring(0, 150) + '...' 
            : contentText;
        
        resultItem.appendChild(content);
        resultsDiv.appendChild(resultItem);
    });
}

function stripHtml(html) {
    return html.replace(/<[^>]*>/g, '').trim();
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('ko-KR');
}

const autocompleteDiv = document.getElementById('autocomplete');
let autocompleteTimeout = null;

// 입력할 때마다 자동완성 호출
searchInput.addEventListener('input', (e) => {
    const query = e.target.value.trim();
    
    // 이전 타이머 취소
    if (autocompleteTimeout) {
        clearTimeout(autocompleteTimeout);
    }
    
    // 2글자 미만이면 숨기기
    if (query.length < 2) {
        autocompleteDiv.classList.remove('show');
        return;
    }
    
    // 300ms 후에 자동완성 요청 (타이핑 멈춤 대기)
    autocompleteTimeout = setTimeout(() => {
        fetchAutocomplete(query);
    }, 300);
});

// 자동완성 데이터 가져오기
async function fetchAutocomplete(query) {
    try {
        const response = await fetch(`${API_BASE_URL}/autocomplete?q=${encodeURIComponent(query)}`);
        
        if (!response.ok) {
            return;
        }

        const data = await response.json();
        displayAutocomplete(data.suggestions, query);
        
    } catch (error) {
        console.error('Autocomplete error:', error);
    }
}

// 자동완성 결과 표시
function displayAutocomplete(suggestions, query) {
    if (!suggestions || suggestions.length === 0) {
        autocompleteDiv.classList.remove('show');
        return;
    }

    autocompleteDiv.innerHTML = suggestions.map(suggestion => {
        // 검색어 하이라이트
        const highlighted = suggestion.replace(
            new RegExp(escapeRegex(query), 'gi'), 
            match => `<strong>${match}</strong>`
        );
        
        return `
            <div class="autocomplete-item" data-value="${escapeHtml(suggestion)}">
                ${highlighted}
            </div>
        `;
    }).join('');
    
    autocompleteDiv.classList.add('show');
    
    // 클릭 이벤트 추가
    autocompleteDiv.querySelectorAll('.autocomplete-item').forEach(item => {
        item.addEventListener('click', () => {
            searchInput.value = item.dataset.value;
            autocompleteDiv.classList.remove('show');
            performSearch();
        });
    });
}

// 외부 클릭 시 자동완성 숨기기
document.addEventListener('click', (e) => {
    if (!searchInput.contains(e.target) && !autocompleteDiv.contains(e.target)) {
        autocompleteDiv.classList.remove('show');
    }
});

// HTML 이스케이프
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 정규식 이스케이프
function escapeRegex(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}