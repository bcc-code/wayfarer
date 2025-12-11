import http from 'k6/http';
import { check } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// Custom metrics for GraphQL-specific tracking
export const graphqlErrors = new Counter('graphql_errors');
export const graphqlDuration = new Trend('graphql_duration');

/**
 * Execute a GraphQL request
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} query - GraphQL query string
 * @param {object} variables - Query variables
 * @param {string} token - JWT token for authorization
 * @param {string} operationName - Optional operation name for tagging
 * @returns {object} HTTP response
 */
export function graphqlRequest(baseUrl, query, variables, token, operationName) {
    const url = `${baseUrl}/graphql`;
    const payload = JSON.stringify({
        query: query,
        variables: variables || {},
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
        tags: {
            name: operationName || 'GraphQL',
        },
    };

    const startTime = Date.now();
    const response = http.post(url, payload, params);
    graphqlDuration.add(Date.now() - startTime, { operation: operationName });

    return response;
}

/**
 * Check if a GraphQL response is successful
 * @param {object} response - HTTP response object
 * @param {string} queryName - Name of the query for logging
 * @returns {boolean} Whether the response was successful
 */
export function checkGraphQLResponse(response, queryName) {
    const checks = {
        [`${queryName}: status is 200`]: (r) => r.status === 200,
        [`${queryName}: has data`]: (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.data !== undefined && body.data !== null;
            } catch {
                return false;
            }
        },
        [`${queryName}: no errors`]: (r) => {
            try {
                const body = JSON.parse(r.body);
                return !body.errors || body.errors.length === 0;
            } catch {
                return false;
            }
        },
    };

    const success = check(response, checks);

    if (!success) {
        graphqlErrors.add(1, { operation: queryName });

        // Log error details for debugging
        if (response.status !== 200) {
            console.error(`${queryName}: HTTP ${response.status}`);
        } else {
            try {
                const body = JSON.parse(response.body);
                if (body.errors) {
                    console.error(`${queryName}: ${JSON.stringify(body.errors)}`);
                }
            } catch {
                console.error(`${queryName}: Invalid JSON response`);
            }
        }
    }

    return success;
}

/**
 * Parse GraphQL response body
 * @param {object} response - HTTP response object
 * @returns {object|null} Parsed response data or null on error
 */
export function parseResponse(response) {
    try {
        const body = JSON.parse(response.body);
        return body.data;
    } catch {
        return null;
    }
}
