import pandas as pd
import re

# Load data
gl_df = pd.read_csv('gl_report.csv', dtype=str)
rtsp_df = pd.read_csv('rtsp_report.csv', dtype=str)

# Clean column names
gl_df.columns = gl_df.columns.str.strip()
rtsp_df.columns = rtsp_df.columns.str.strip()

# Step 1: Extract transaction ID from GL system narration
def extract_txn_id(narration):
    match = re.search(r'/([A-Za-z0-9]{10,})/', str(narration))
    return match.group(1) if match else None

gl_df['EXTRACTED_TXN_ID'] = gl_df['SYSTEM_NARRATION'].apply(extract_txn_id)

# Step 2: Filter RTSP to exclude "Inward Transfer" and "Outward Transfer"
exclude_types = ['Inward Transfer', 'Outward Transfer']
rtsp_df_filtered = rtsp_df[~rtsp_df['Type Of Transaction'].str.strip().isin(exclude_types)]

# Step 3: Compare by Transaction ID
gl_txn_ids = set(gl_df['EXTRACTED_TXN_ID'].dropna())
rtsp_txn_ids = set(rtsp_df_filtered['Transaction ID'].dropna())

# Transactions in RTSP not in GL
only_in_rtsp = rtsp_df_filtered[~rtsp_df_filtered['Transaction ID'].isin(gl_txn_ids)]

# Transactions in GL not in RTSP
only_in_gl = gl_df[~gl_df['EXTRACTED_TXN_ID'].isin(rtsp_txn_ids)]

# Step 4: Save output to CSV
only_in_rtsp.to_csv('missing_in_gl.csv', index=False)
only_in_gl.to_csv('missing_in_rtsp.csv', index=False)

print("Comparison complete. Files generated: 'missing_in_gl.csv' and 'missing_in_rtsp.csv'")
