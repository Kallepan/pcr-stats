import os
import pandas as pd
import logging
import re
from typing import List

logging.basicConfig(level=logging.INFO)

OUTPUT_PATH = os.path.join(os.path.dirname(__file__), "output")

if not os.path.exists(OUTPUT_PATH):
    os.makedirs(OUTPUT_PATH)


FILE_PATH = os.path.join(os.path.dirname(__file__), "data")
CODES_TO_BE_FILTERED = [
    "20042",
    "20039",
    "20009",
    "20008",
    "20045",
    "20031",
    "2506",
    "20034",
    "20005",
    "20045",
    "20012",
    "20043",
    "20040",
    "2002",
    "20054",
    "2006",
    "9007",
    "20055",
    "20041",
    "20038",
    "20025",
    "20013",
]


def get_files(path: str) -> list:
    files = os.listdir(path)
    return files


def create_statistics(df: pd.DataFrame) -> pd.DataFrame:
    # count all occurence of each cod
    statistics = df["code"].value_counts().reset_index()
    statistics.columns = ["code", "count"]
    statistics = statistics[statistics["code"].isin(CODES_TO_BE_FILTERED)]
    return statistics


def main():
    files = get_files(FILE_PATH)

    # create a df
    line_regex = re.compile(r"(\d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}.\d{3})")
    all_lines_collection: List[str] = []
    df = pd.DataFrame()
    for file in files:
        logging.info(f"Processing {file}")
        # read file line by line
        with open(os.path.join(FILE_PATH, file), "r") as f:
            lines = []

            for i, line in enumerate(f.readlines()):
                if line is None:
                    continue
                if line == "\n" or line.strip() == "":
                    continue

                if line_regex.match(line) is None:
                    lines[-1] = lines[-1].strip() + " " + line.strip().replace("\n", "")
                else:
                    lines.append(line.strip().replace("\n", ""))

            all_lines_collection.extend(lines)

    # cleanup
    all_lines_collection = [
        row.replace("\t\t", "\t").split("\t") for row in all_lines_collection
    ]
    # fetch only the first, second, third, fourth and last column
    all_lines_collection = [
        [row[0], row[1], row[2], row[3], row[-1]] for row in all_lines_collection
    ]
    with open(os.path.join(OUTPUT_PATH, "all_lines.txt"), "w") as f:
        for line in all_lines_collection:
            f.write("\t".join(line) + "\n")

    # create a dataframe
    df = pd.DataFrame(
        all_lines_collection, columns=["date", "user", "type", "code", "message"]
    )

    # generate statistics
    statistics = create_statistics(df)

    # get the message from the first occurence of each code
    first_occurence = df.drop_duplicates(subset="code", keep="first")

    # merge the first occurence and the statistics
    statistics = pd.merge(first_occurence, statistics, on="code")
    statistics.to_csv(os.path.join(OUTPUT_PATH, "statistics.csv"), index=False)


if __name__ == "__main__":
    main()
