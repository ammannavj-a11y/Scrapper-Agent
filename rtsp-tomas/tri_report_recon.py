import pandas as pd
import re

def normalize_status(value):
    return str(value).strip().lower()

def extract_txn_id_from_gl(narration):
    """Extracts numeric txn ID from GL SYSTEM_NARRATION field"""
    match = re.search(r'/(\d{10,})/', str(narration))
    return match.group(1) if match else None

def process_all_reports(rtsp_file, tranlog_file, gl_file, output_base):
    # === Load RTSP ===
    rtsp_df = pd.read_csv(rtsp_file, dtype=str)
    rtsp_df.columns = rtsp_df.columns.str.strip()
    rtsp_df['Transaction ID'] = rtsp_df['Transaction ID'].str.strip()
    rtsp_df['STATUS'] = rtsp_df['STATUS'].apply(normalize_status)
    rtsp_dict = rtsp_df.set_index('Transaction ID')['STATUS'].to_dict()

    # === Load TRANLOG ===
    tran_df = pd.read_csv(tranlog_file, dtype=str)
    tran_df.columns = tran_df.columns.str.strip()
    tran_df['TXN_ID'] = tran_df['TXN_ID'].str.strip()
    tran_df['STATUS'] = tran_df['STATUS'].apply(normalize_status)
    tranlog_dict = tran_df.set_index('TXN_ID')['STATUS'].to_dict()

    # === Load GL ===
    gl_df = pd.read_csv(gl_file, dtype=str)
    gl_df.columns = gl_df.columns.str.strip()
    gl_df['TXN_ID'] = gl_df['SYSTEM_NARRATION'].apply(extract_txn_id_from_gl)
    gl_df = gl_df.dropna(subset=['TXN_ID'])
    gl_txn_ids = set(gl_df['TXN_ID'].unique())

    rtsp_txns = set(rtsp_dict.keys())
    tranlog_txns = set(tranlog_dict.keys())

    # ✅ 1. In RTSP + TOMAS but NOT in GL
    in_rtsp_tranlog = rtsp_txns & tranlog_txns
    missing_in_gl = [
        txn_id for txn_id in in_rtsp_tranlog if txn_id not in gl_txn_ids
    ]

    df_missing_gl = pd.DataFrame([{
        "TXN_ID": txn_id,
        "RTSP_STATUS": rtsp_dict.get(txn_id),
        "TRANLOG_STATUS": tranlog_dict.get(txn_id)
    } for txn_id in missing_in_gl])

    # ✅ 2. GL + RTSP pending + TOMAS success
    gl_rtsp_pending_tranlog_success = []
    for txn_id in gl_txn_ids:
        rtsp_status = rtsp_dict.get(txn_id)
        tran_status = tranlog_dict.get(txn_id)
        if rtsp_status == 'pending' and tran_status == 'success':
            gl_rtsp_pending_tranlog_success.append({
                "TXN_ID": txn_id,
                "RTSP_STATUS": rtsp_status,
                "TRANLOG_STATUS": tran_status
            })

    df_gl_rtsp_pending_tomas_success = pd.DataFrame(gl_rtsp_pending_tranlog_success)

    # ✅ 3. GL + TOMAS pending + RTSP success
    gl_tomas_pending_rtsp_success = []
    for txn_id in gl_txn_ids:
        rtsp_status = rtsp_dict.get(txn_id)
        tran_status = tranlog_dict.get(txn_id)
        if tran_status == 'pending' and rtsp_status == 'success':
            gl_tomas_pending_rtsp_success.append({
                "TXN_ID": txn_id,
                "RTSP_STATUS": rtsp_status,
                "TRANLOG_STATUS": tran_status
            })

    df_gl_tomas_pending_rtsp_success = pd.DataFrame(gl_tomas_pending_rtsp_success)

    # === Write to Excel ===
    excel_output_file = f"{output_base}.xlsx"
    with pd.ExcelWriter(excel_output_file, engine='xlsxwriter') as writer:
        df_missing_gl.to_excel(writer, index=False, sheet_name="MissingInGL")
        df_gl_rtsp_pending_tomas_success.to_excel(writer, index=False, sheet_name="GL_RTSP_Pending_TOMAS_Success")
        df_gl_tomas_pending_rtsp_success.to_excel(writer, index=False, sheet_name="GL_TOMAS_Pending_RTSP_Success")

    print(f"Output written to {excel_output_file}")

# Usage
if __name__ == "__main__":
    process_all_reports("rtsp_report.csv", "tranlog_report.csv", "gl_report.csv", "tri_recon_output")
