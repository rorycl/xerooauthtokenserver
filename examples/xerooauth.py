"""
Simple Xero Oauth2 flow based heavily on
https://edgecate.com/articles/how-to-access-xero-apis/
"""

# import json
import requests
import base64
import os
import sys

client_id = os.getenv("XEROCLIENTID")
client_secret = os.getenv("XEROCLIENTSECRET")
redirect_url = 'https://xero.com/'
scope = 'offline_access accounting.transactions'
b64_id_secret = base64.b64encode(bytes(
        client_id + ':' + client_secret, 'utf-8')
    ).decode('utf-8')

if not client_id or not client_secret:
    print("client id and/or secret not in environment, exiting")
    sys.exit(1)


def XeroFirstAuth():
    # generate the auth url
    auth_url = ('''https://login.xero.com/identity/connect/authorize?''' +
                '''response_type=code''' +
                '''&client_id=''' + client_id +
                '''&redirect_uri=''' + redirect_url +
                '''&scope=''' + scope +
                '''&state=123''')
    print("Please go to the following url and then paste the response " +
          "url here after authenticating\n" + auth_url)

    # extract code from response code
    auth_res_url = input('What is the response URL? ')
    start_number = auth_res_url.find('code=') + len('code=')
    end_number = auth_res_url.find('&scope')
    auth_code = auth_res_url[start_number:end_number]
    print(auth_code)
    print('\n')

    # retrieve a token using the code
    exchange_code_url = 'https://identity.xero.com/connect/token'
    response = requests.post(
        exchange_code_url,
        headers={
            'Authorization': 'Basic ' + b64_id_secret
        },
        data={
            'grant_type': 'authorization_code',
            'code': auth_code,
            'redirect_uri': redirect_url
        })
    json_response = response.json()
    print(json_response)
    print('\n')

    rt_file = open('token.txt', 'w')
    rt_file.write(json_response['access_token'])
    rt_file.close()

    rt_file = open('refresh_token.txt', 'w')
    rt_file.write(json_response['refresh_token'])
    rt_file.close()

    # retrieve token and refresh token from json response
    return [json_response['access_token'], json_response['refresh_token']]


# Check tenants
def XeroTenants(access_token):
    connections_url = 'https://api.xero.com/connections'
    response = requests.get(
        connections_url,
        headers={
            'Authorization': 'Bearer ' + access_token,
            'Content-Type': 'application/json'
        })
    json_response = response.json()
    print(json_response)

    for tenants in json_response:
        json_dict = tenants
    return json_dict['tenantId']


# Refresh access token
def XeroRefreshToken(refresh_token):
    token_refresh_url = 'https://identity.xero.com/connect/token'
    response = requests.post(
        token_refresh_url,
        headers={
            'Authorization': 'Basic ' + b64_id_secret,
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        data={
            'grant_type': 'refresh_token',
            'refresh_token': refresh_token
        })
    json_response = response.json()
    print(json_response)

    new_refresh_token = json_response['refresh_token']
    rt_file = open('refresh_token.txt', 'w')
    rt_file.write(new_refresh_token)
    rt_file.close()

    print("access_token %s\nrefresh token %s\n" % (
        json_response['access_token'], json_response['refresh_token']))
    return [json_response['access_token'], json_response['refresh_token']]


def XeroRequests(access_token=None, tenant_id=None):
    if not access_token or not tenant_id:
        old_refresh_token = open('refresh_token.txt', 'r').read()
        access_token, refresh_token = XeroRefreshToken(old_refresh_token)
        tenant_id = XeroTenants(access_token)

    get_url = 'https://api.xero.com/api.xro/2.0/Invoices'
    response = requests.get(
        get_url,
        headers={
            'Authorization': 'Bearer ' + access_token,
            'Xero-tenant-id': tenant_id,
            'Accept': 'application/json'
        })
    json_response = response.json()
    print(json_response)


if __name__ == '__main__':

    access_token, refresh_token = XeroFirstAuth()
    tenants = XeroTenants(access_token)
    XeroRequests()
