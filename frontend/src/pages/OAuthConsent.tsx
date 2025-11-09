import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import axios from 'axios';
import { useAuth } from '../context/AuthContext';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export const OAuthConsent: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { token, isAuthenticated } = useAuth();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [consentData, setConsentData] = useState<any>(null);

  const clientId = searchParams.get('client_id');
  const redirectUri = searchParams.get('redirect_uri');
  const scope = searchParams.get('scope');
  const state = searchParams.get('state');
  const responseType = searchParams.get('response_type');

  useEffect(() => {
    if (!isAuthenticated) {
      // Redirect to login with return URL
      const returnUrl = `/oauth/consent?${searchParams.toString()}`;
      navigate(`/login?return=${encodeURIComponent(returnUrl)}`);
      return;
    }

    if (!clientId || !redirectUri || responseType !== 'code') {
      setError('Invalid OAuth request parameters');
      setLoading(false);
      return;
    }

    // Fetch client information
    const fetchConsentData = async () => {
      try {
        const response = await axios.get(
          `${API_URL}/oauth/authorize?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=${responseType}&scope=${scope || ''}&state=${state || ''}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );
        setConsentData(response.data);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to load consent information');
      } finally {
        setLoading(false);
      }
    };

    fetchConsentData();
  }, [isAuthenticated, clientId, redirectUri, responseType, scope, state, token, navigate, searchParams]);

  const handleConsent = async (approved: boolean) => {
    setLoading(true);
    setError('');

    try {
      const response = await axios.post(
        `${API_URL}/oauth/authorize/consent`,
        {
          client_id: clientId,
          redirect_uri: redirectUri,
          scope: scope || '',
          state: state || '',
          approved,
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      // Redirect to the application
      window.location.href = response.data.redirect_url;
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to process consent');
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-md w-full">
          <div className="bg-white shadow-md rounded-lg p-6">
            <div className="text-center">
              <h2 className="text-2xl font-bold text-red-600 mb-4">Error</h2>
              <p className="text-gray-700">{error}</p>
              <button
                onClick={() => navigate('/dashboard')}
                className="mt-4 px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
              >
                Go to Dashboard
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full">
        <div className="bg-white shadow-md rounded-lg overflow-hidden">
          <div className="bg-indigo-600 px-6 py-4">
            <h2 className="text-2xl font-bold text-white text-center">Authorization Request</h2>
          </div>

          <div className="px-6 py-8">
            <div className="text-center mb-6">
              <p className="text-lg text-gray-700">
                <span className="font-semibold">{consentData?.client_name || 'An application'}</span> is requesting
                access to your account.
              </p>
            </div>

            {scope && (
              <div className="mb-6">
                <h3 className="text-sm font-medium text-gray-700 mb-2">This application will be able to:</h3>
                <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                  {scope.split(' ').map((s, i) => (
                    <li key={i}>{s}</li>
                  ))}
                </ul>
              </div>
            )}

            <div className="bg-gray-50 rounded-md p-4 mb-6">
              <p className="text-xs text-gray-600">
                By authorizing this application, you allow it to access your account information on your behalf.
                You can revoke access at any time from your dashboard.
              </p>
            </div>

            <div className="flex space-x-4">
              <button
                onClick={() => handleConsent(false)}
                disabled={loading}
                className="flex-1 py-2 px-4 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
              >
                Deny
              </button>
              <button
                onClick={() => handleConsent(true)}
                disabled={loading}
                className="flex-1 py-2 px-4 border border-transparent rounded-md text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
              >
                Authorize
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
