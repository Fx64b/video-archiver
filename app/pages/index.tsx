import * as React from "react";
import {Button} from '@nextui-org/react'
import { Card, Title, DonutChart } from "@tremor/react";

const cities = [
    {
        name: "New York",
        sales: 9800,
    },
    {
        name: "London",
        sales: 4567,
    },
    {
        name: "Hong Kong",
        sales: 3908,
    },
    {
        name: "San Francisco",
        sales: 2400,
    },
    {
        name: "Singapore",
        sales: 1908,
    },
    {
        name: "Zurich",
        sales: 1398,
    },
];

const valueFormatter = (number: number) => `$ ${Intl.NumberFormat("us").format(number).toString()}`;

export default function Home() {
    return (
    <main className={'flex flex-col justify-center items-center'}>
        <h1 className={'text-5xl text-center mt-10'}>Next.js + TypeScript + NextUI + Tremor + Chart.js</h1>
        <Button color="primary" className={'mt-10'}>Click Me!</Button>
        <Card className="max-w-lg mt-10">
            <Title>Sales</Title>
            <DonutChart
                className="mt-6"
                data={cities}
                category="sales"
                index="name"
                valueFormatter={valueFormatter}
                colors={["slate", "violet", "indigo", "rose", "cyan", "amber"]}
            />
        </Card>
    </main>
  )
}
