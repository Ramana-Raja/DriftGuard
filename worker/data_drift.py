import pandas as pd
import numpy as np

from scipy.stats import (
    ks_2samp,
    wasserstein_distance,
    chi2_contingency,
    entropy,
)


def calculate_psi(expected, actual, bins=10):

    expected = np.array(expected)
    actual = np.array(actual)

    breakpoints = np.linspace(0, 100, bins + 1)
    breakpoints = np.unique(np.percentile(expected, breakpoints))

    expected_counts = np.histogram(expected, bins=breakpoints)[0]
    actual_counts = np.histogram(actual, bins=breakpoints)[0]

    expected_perc = expected_counts / len(expected)
    actual_perc = actual_counts / len(actual)

    expected_perc = np.where(expected_perc == 0, 0.0001, expected_perc)
    actual_perc = np.where(actual_perc == 0, 0.0001, actual_perc)

    psi = np.sum(
        (actual_perc - expected_perc)
        * np.log(actual_perc / expected_perc)
    )

    return float(psi)


def calculate_js_divergence(expected, actual, bins=10):

    expected_hist, bin_edges = np.histogram(
        expected,
        bins=bins,
        density=True,
    )

    actual_hist, _ = np.histogram(
        actual,
        bins=bin_edges,
        density=True,
    )

    expected_hist = expected_hist + 1e-10
    actual_hist = actual_hist + 1e-10

    m = 0.5 * (expected_hist + actual_hist)

    js = 0.5 * (
        entropy(expected_hist, m)
        + entropy(actual_hist, m)
    )

    return float(js)


def check_data_drift(
    ref_df,
    cur_df,
    p_threshold=0.05,
):


    common_cols = list(
        set(ref_df.columns).intersection(
            set(cur_df.columns)
        )
    )

    numeric_cols = list(
        set(
            ref_df[common_cols]
            .select_dtypes(include=np.number)
            .columns
        ).intersection(
            set(
                cur_df[common_cols]
                .select_dtypes(include=np.number)
                .columns
            )
        )
    )

    categorical_cols = list(
        set(common_cols) - set(numeric_cols)
    )

    results = {
        "ks_test": {},
        "psi": {},
        "wasserstein_distance": {},
        "chi_square": {},
        "correlation_drift": {},
    }

    for col in numeric_cols:

        ref_series = ref_df[col].dropna()
        cur_series = cur_df[col].dropna()

        if len(ref_series) == 0 or len(cur_series) == 0:
            continue

        ks_stat, ks_pvalue = ks_2samp(
            ref_series,
            cur_series,
        )

        results["ks_test"][col] = {
            "statistic": round(float(ks_stat), 6),
            "p_value": round(float(ks_pvalue), 6),
            "drift_detected": ks_pvalue < p_threshold,
        }

        psi_score = calculate_psi(
            ref_series,
            cur_series,
        )

        results["psi"][col] = {
            "score": round(float(psi_score), 6),
            "drift_detected": psi_score > 0.25,
        }


        wasserstein_score = wasserstein_distance(
            ref_series,
            cur_series,
        )

        results["wasserstein_distance"][col] = {
            "score": round(float(wasserstein_score), 6),
        }

    for col in categorical_cols:

        ref_counts = (
            ref_df[col]
            .fillna("NULL")
            .value_counts()
        )

        cur_counts = (
            cur_df[col]
            .fillna("NULL")
            .value_counts()
        )

        categories = list(
            set(ref_counts.index).union(
                set(cur_counts.index)
            )
        )

        ref_freq = [
            ref_counts.get(cat, 0)
            for cat in categories
        ]

        cur_freq = [
            cur_counts.get(cat, 0)
            for cat in categories
        ]

        contingency_table = np.array(
            [ref_freq, cur_freq]
        )

        try:

            chi2_stat, chi2_pvalue, _, _ = (
                chi2_contingency(
                    contingency_table
                )
            )

            results["chi_square"][col] = {
                "statistic": round(
                    float(chi2_stat),
                    6,
                ),
                "p_value": round(
                    float(chi2_pvalue),
                    6,
                ),
                "drift_detected": (
                    chi2_pvalue < p_threshold
                ),
            }

        except Exception as e:

            results["chi_square"][col] = {
                "error": str(e)
            }

    if len(numeric_cols) > 1:

        ref_corr = (
            ref_df[numeric_cols]
            .corr()
        )

        cur_corr = (
            cur_df[numeric_cols]
            .corr()
        )

        corr_diff = (
            ref_corr - cur_corr
        ).abs()

        results["correlation_drift"] = (
            corr_diff.to_dict()
        )

    overall_drift = False
    for col_metrics in results["ks_test"].values():

        if col_metrics["drift_detected"]:
            overall_drift = True
            break
    if not overall_drift:

        for col_metrics in results["psi"].values():

            if col_metrics["drift_detected"]:
                overall_drift = True
                break
    if not overall_drift:

        for col_metrics in results["chi_square"].values():

            if (
                    isinstance(col_metrics, dict)
                    and col_metrics.get(
                "drift_detected",
                False,
            )
            ):
                overall_drift = True
                break

    if not overall_drift:

        corr_matrix = results["correlation_drift"]

        for feature_1 in corr_matrix:

            for feature_2 in corr_matrix[feature_1]:

                corr_change = corr_matrix[
                    feature_1
                ][feature_2]
                if corr_change > 0.2:
                    overall_drift = True
                    break

    results["data_drift_detected"] = (
        overall_drift
    )
    return results

# import pandas as pd
# data_ref = pd.read_csv("iris_dataset.csv")
# data_curr = pd.read_csv("iris_dataset.csv")
# print(check_data_drift(data_ref,data_curr))