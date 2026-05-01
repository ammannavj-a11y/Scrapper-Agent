import pandas as pd

def normalize_status(value):
    return str(value).strip().lower()

def process_reports(rtsp_file, tranlog_file, outfile):
    # Load RTSP data
    #rtsp_df = pd.read_csv(rtsp_file, dtype=str)
    #rtsp_df['Transaction ID'] = rtsp_df['Transaction ID'].str.strip()
    #rtsp_df['STATUS'] = rtsp_df['STATUS'].apply(normalize_status)

    # Load RTSP data
    rtsp_df = pd.read_csv(rtsp_file, dtype=str)
    rtsp_df.columns = rtsp_df.columns.str.strip()  # 🧽 Clean header names
    print("Actual columns in RTSP file:", rtsp_df.columns.tolist())

    rtsp_df['Transaction ID'] = rtsp_df['Transaction ID'].str.strip()
    rtsp_df['STATUS'] = rtsp_df['STATUS'].apply(normalize_status)


    print("RTSP Columns:", rtsp_df.columns.tolist())

    rtsp_dict = rtsp_df.set_index('Transaction ID')['STATUS'].to_dict()

    # Load TRANLOG data
    tran_df = pd.read_csv(tranlog_file, dtype=str)
    tran_df['TXN_ID'] = tran_df['TXN_ID'].str.strip()
    tran_df['STATUS'] = tran_df['STATUS'].apply(normalize_status)

    # Output results list
    output_records = []

    for index, row in tran_df.iterrows():
        txn_id = row['TXN_ID']
        tran_status = row['STATUS']
        rtsp_status = rtsp_dict.get(txn_id)

        if rtsp_status is None:
            condition = "MISSING_IN_RTSP"
        elif tran_status == 'c' and rtsp_status != 'success':
            condition = "SUCCESS_MISMATCH"
        elif tran_status != 'c' and rtsp_status == 'success':
            condition = "SUCCESS_MISMATCH"
        elif tran_status == 'c' and rtsp_status == 'success':
            condition = "MATCH_SUCCESS"
        else:
            condition = "MATCH_FAILURE"

        output_records.append({
            "TXN_ID": txn_id,
            "PresentInRTSP": "No" if rtsp_status is None else "Yes",
            "RTSP_STATUS": rtsp_status if rtsp_status else "N/A",
            "TRAN_STATUS": tran_status,
            "STATUS_CONDITION": condition
        })

    out_df = pd.DataFrame(output_records)

    # Write to Excel with condition-based sheets
    with pd.ExcelWriter(outfile, engine='xlsxwriter') as writer:
        out_df.to_excel(writer, index=False, sheet_name="All_Results")
        out_df[out_df["STATUS_CONDITION"] == "MISSING_IN_RTSP"].to_excel(writer, index=False, sheet_name="MissingInRTSP")
        out_df[out_df["STATUS_CONDITION"] == "SUCCESS_MISMATCH"].to_excel(writer, index=False, sheet_name="Mismatch")

    print(f"Comparison complete. Excel output written to {outfile}")

# Usage
if __name__ == "__main__":
    process_reports("rtsp_report.csv", "tranlog_report.csv", "outfile.xlsx")
