"""
Python xerooauthtokenserver integration example
"""

import requests


def get_token():
    """retrieve token from xerooauthtoken server"""
    response = requests.get("http://127.0.0.1:5001/token")
    return response.json()['accessToken']


def tenants(access_token):
    """retrieve first tenant id"""
    tenants_url = 'https://api.xero.com/connections'
    response = requests.get(
        tenants_url,
        headers={
            'Authorization': 'Bearer ' + access_token,
            'Content-Type': 'application/json'
        })
    return response.json()[0]["tenantId"]  # first tenant id


def invoices(access_token, tenant_id):
    """get invoices"""
    invoice_url = 'https://api.xero.com/api.xro/2.0/Invoices'
    response = requests.get(
        invoice_url,
        headers={
            'Authorization': 'Bearer ' + access_token,
            'Xero-tenant-id': tenant_id,
            'Accept': 'application/json'
        })
    return response.json()


if __name__ == '__main__':

    token = get_token()
    tenant = tenants(token)
    invoices = invoices(token, tenant)
    print(invoices)
