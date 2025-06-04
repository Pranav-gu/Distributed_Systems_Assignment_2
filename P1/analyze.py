import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import sys
import os
from datetime import datetime

def analyze_load_test(csv_file):
    # Load data
    print(f"Loading data from {csv_file}...")
    df = pd.read_csv(csv_file)
    
    # Convert time strings to datetime
    df['StartTime'] = pd.to_datetime(df['StartTime'])
    df['EndTime'] = pd.to_datetime(df['EndTime'])
    
    # Convert duration to numeric if it's not already
    df['Duration_ms'] = pd.to_numeric(df['Duration_ms'])
    
    # Create output directory
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_dir = f"analysis_results_{timestamp}"
    os.makedirs(output_dir, exist_ok=True)
    
    # Basic statistics
    total_requests = len(df)
    successful_requests = df['Success'].sum()
    success_rate = (successful_requests / total_requests) * 100
    
    avg_duration = df[df['Success'] == True]['Duration_ms'].mean()
    median_duration = df[df['Success'] == True]['Duration_ms'].median()
    max_duration = df[df['Success'] == True]['Duration_ms'].max()
    min_duration = df[df['Success'] == True]['Duration_ms'].min()
    
    # Calculate test duration
    test_duration_seconds = (df['EndTime'].max() - df['StartTime'].min()).total_seconds()
    throughput = total_requests / test_duration_seconds
    
    # Summary report
    summary = f"""
    Load Test Analysis Summary
    =========================
    Total Requests: {total_requests}
    Successful Requests: {successful_requests} ({success_rate:.2f}%)
    Test Duration: {test_duration_seconds:.2f} seconds
    
    Response Time Statistics:
    ------------------------
    Average: {avg_duration:.2f} ms
    Median: {median_duration:.2f} ms
    Min: {min_duration:.2f} ms
    Max: {max_duration:.2f} ms
    
    Throughput: {throughput:.2f} requests/second
    """
    
    print(summary)
    
    with open(f"{output_dir}/summary.txt", "w") as f:
        f.write(summary)
    
    # Extract server load distribution
    server_distribution = df['TargetServer'].value_counts()
    server_distribution_pct = (server_distribution / total_requests) * 100
    
    # Time series of requests
    df['TimePoint'] = df['StartTime'].astype(np.int64) // 10**9
    min_time = df['TimePoint'].min()
    df['RelativeSecond'] = df['TimePoint'] - min_time
    
    # Group by time bucket (1-second intervals)
    requests_per_second = df.groupby('RelativeSecond').size()
    
    # Response time over time
    response_time_over_time = df.groupby('RelativeSecond')['Duration_ms'].mean()
    
    # ===== Visualizations =====
    
    # 1. Response Time Distribution
    plt.figure(figsize=(10, 6))
    plt.hist(df[df['Success'] == True]['Duration_ms'], bins=30, alpha=0.7)
    plt.xlabel('Response Time (ms)')
    plt.ylabel('Number of Requests')
    plt.title('Response Time Distribution')
    plt.grid(True, alpha=0.3)
    plt.savefig(f"{output_dir}/response_time_distribution.png", dpi=300)
    
    # 2. Server Load Distribution
    plt.figure(figsize=(12, 6))
    server_distribution_pct.plot(kind='bar', color='skyblue')
    plt.xlabel('Server')
    plt.ylabel('Percentage of Requests (%)')
    plt.title('Load Distribution Across Servers')
    plt.xticks(rotation=45)
    plt.tight_layout()
    plt.grid(True, axis='y', alpha=0.3)
    plt.savefig(f"{output_dir}/server_load_distribution.png", dpi=300)
    
    # 3. Requests Per Second
    plt.figure(figsize=(12, 6))
    requests_per_second.plot(kind='line', marker='o', color='green', alpha=0.7)
    plt.xlabel('Time (seconds since start)')
    plt.ylabel('Number of Requests')
    plt.title('Requests Per Second')
    plt.grid(True, alpha=0.3)
    plt.savefig(f"{output_dir}/requests_per_second.png", dpi=300)
    
    # 4. Response Time Over Time
    plt.figure(figsize=(12, 6))
    response_time_over_time.plot(kind='line', marker='o', color='red', alpha=0.7)
    plt.xlabel('Time (seconds since start)')
    plt.ylabel('Average Response Time (ms)')
    plt.title('Average Response Time Over Time')
    plt.grid(True, alpha=0.3)
    plt.savefig(f"{output_dir}/response_time_over_time.png", dpi=300)
    
    # 5. Client Response Time Box Plot
    plt.figure(figsize=(12, 6))
    client_data = df.groupby('ClientID')['Duration_ms'].agg(['median', 'mean', 'max', 'min'])
    client_data['median'].plot(kind='box', vert=False)
    plt.xlabel('Response Time (ms)')
    plt.title('Response Time Distribution Across Clients')
    plt.grid(True, axis='x', alpha=0.3)
    plt.savefig(f"{output_dir}/client_response_time.png", dpi=300)
    
    print(f"\nAnalysis complete. Results saved to '{output_dir}' directory.")
    return output_dir

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python analyze_results.py <csv_file>")
        sys.exit(1)
    
    csv_file = sys.argv[1]
    analyze_load_test(csv_file)