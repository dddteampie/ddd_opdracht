
// Map servicenamen naar hun omgevingsvariabelen
const API_URLS = {
  behoefte: process.env.REACT_APP_BEHOEFTE_API_BASE_URL,
  aanvraag: process.env.REACT_APP_AANVRAAG_API_BASE_URL,
  product: process.env.REACT_APP_PRODUCT_API_BASE_URL,
  ecd: process.env.REACT_APP_ECD_API_BASE_URL, // Optioneel, afhankelijk van je setup
};

// Functie om de juiste basis-URL op te halen
const getApiBaseUrl = (serviceName) => {
  const url = API_URLS[serviceName];
  if (!url) {
    console.error(`ERROR: API base URL for service '${serviceName}' is not defined.`);
    // Werp een fout of retourneer een fallback URL
    throw new Error(`API base URL for service '${serviceName}' is not configured.`);
  }
  return url;
};

const callApi = async (serviceName, endpoint, options = {}) => {
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  const baseUrl = getApiBaseUrl(serviceName); // Haal de service-specifieke URL op

  const response = await fetch(`${baseUrl}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let errorDetail = await response.text();
    try {
      const errorJson = JSON.parse(errorDetail);
      errorDetail = errorJson.message || errorDetail;
    } catch (e) {
      // ignore
    }
    throw new Error(`API Error ${response.status}: ${errorDetail}`);
  }

  if (response.status === 204 || response.headers.get('Content-Length') === '0') {
    return null;
  }
  
  return response.json();
};

export default callApi;