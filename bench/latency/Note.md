
Running on Mac M3 with 8GB RAM ()

`./run.sh`:

```go
large.jpg      direct    p50=   0.28ms p99=   0.75ms errors=0
large.jpg      raw       p50=  12.07ms p99=  19.34ms errors=0
large.jpg      reencode  p50=  70.13ms p99= 118.99ms errors=0
large.jpg      resize    p50=  23.71ms p99=  31.57ms errors=0
large.jpg      webp      p50=  27.57ms p99=  39.52ms errors=0
large.jpg      avif      p50=  25.55ms p99=  27.94ms errors=0
medium.jpg     direct    p50=   0.25ms p99=   1.64ms errors=0
medium.jpg     raw       p50=   2.49ms p99=   7.81ms errors=0
medium.jpg     reencode  p50=  14.13ms p99=  28.49ms errors=0
medium.jpg     resize    p50=   6.18ms p99=   9.33ms errors=0
medium.jpg     webp      p50=   9.01ms p99=  11.85ms errors=0
medium.jpg     avif      p50=   7.53ms p99=   9.33ms errors=0
medium.png     direct    p50=   0.38ms p99=   1.69ms errors=0
medium.png     raw       p50=  14.79ms p99=  18.42ms errors=0
medium.png     reencode  p50= 118.53ms p99= 197.38ms errors=0
medium.png     resize    p50=  28.32ms p99= 110.51ms errors=0
medium.png     webp      p50=  31.19ms p99= 129.77ms errors=0
medium.png     avif      p50=  27.48ms p99=  32.86ms errors=0
small.jpg      direct    p50=   0.15ms p99=   0.71ms errors=0
small.jpg      raw       p50=   0.64ms p99=   2.55ms errors=0
small.jpg      reencode  p50=   3.16ms p99=   4.56ms errors=0
small.jpg      resize    p50=   2.69ms p99=   3.93ms errors=0
small.jpg      webp      p50=   7.24ms p99=  12.07ms errors=0
small.jpg      avif      p50=   4.86ms p99=  15.26ms errors=0

Latency (total time incl. body read), 200 requests, concurrency 1

       image  scenario    bytes  ttfb p50     p50     p90     p99      max   rps       p50 vs direct  errors
   large.jpg    direct    2.4MB      0.05    0.28    0.59    0.75     1.09  2995                   -       0
   large.jpg       raw    2.4MB      1.15   12.07   14.47   19.34    48.90    80    +11.79ms (42.7x)       0
   large.jpg  reencode    1.9MB     62.08   70.13   75.93  118.99   190.96    14   +69.85ms (247.8x)       0
   large.jpg    resize    3.9KB     23.64   23.71   24.91   31.57    86.63    41    +23.42ms (83.8x)       0
   large.jpg      webp    2.0KB     27.50   27.57   29.60   39.52    53.98    35    +27.29ms (97.4x)       0
   large.jpg      avif    2.9KB     25.50   25.55   26.87   27.94    30.52    39    +25.26ms (90.3x)       0
  medium.jpg    direct  423.7KB      0.10    0.25    0.53    1.64     3.60  2898                   -       0
  medium.jpg       raw  423.7KB      0.88    2.49    4.33    7.81    11.81   344      +2.24ms (9.9x)       0
  medium.jpg  reencode  346.5KB     13.00   14.13   15.89   28.49    78.75    67    +13.88ms (56.1x)       0
  medium.jpg    resize    3.5KB      6.14    6.18    6.78    9.33     9.95   157     +5.93ms (24.5x)       0
  medium.jpg      webp    1.8KB      8.96    9.01    9.87   11.85    12.73   109     +8.76ms (35.8x)       0
  medium.jpg      avif    2.7KB      7.47    7.53    7.90    9.33    12.15   131     +7.27ms (29.9x)       0
  medium.png    direct    3.2MB      0.05    0.38    0.96    1.69     2.28  1893                   -       0
  medium.png       raw    3.2MB      1.17   14.79   16.90   18.42    19.61    67    +14.41ms (39.0x)       0
  medium.png  reencode    3.0MB    106.42  118.53  132.83  197.38   325.79     8  +118.16ms (312.8x)       0
  medium.png    resize   60.4KB     28.25   28.32   34.26  110.51  1641.87    26    +27.94ms (74.7x)       0
  medium.png      webp    1.8KB     31.14   31.19   42.18  129.77   242.00    27    +30.81ms (82.3x)       0
  medium.png      avif    2.6KB     27.41   27.48   29.61   32.86  1635.21    28    +27.10ms (72.5x)       0
   small.jpg    direct   65.7KB      0.11    0.15    0.28    0.71     0.91  5338                   -       0
   small.jpg       raw   65.7KB      0.55    0.64    1.54    2.55    11.21  1149      +0.49ms (4.3x)       0
   small.jpg  reencode   54.4KB      3.11    3.16    3.48    4.56     5.34   307     +3.02ms (21.5x)       0
   small.jpg    resize    6.8KB      2.65    2.69    2.90    3.93     4.29   364     +2.55ms (18.3x)       0
   small.jpg      webp    3.3KB      7.21    7.24    7.75   12.07    62.85   130     +7.09ms (49.3x)       0
   small.jpg      avif    4.7KB      4.83    4.86    6.11   15.26    43.31   180     +4.72ms (33.1x)       0
```

`./run.sh -n 500 -c 8`

```go
large.jpg      direct    p50=   1.59ms p99=   4.05ms errors=0
large.jpg      raw       p50=  95.25ms p99= 141.16ms errors=0
large.jpg      reencode  p50= 235.95ms p99= 638.50ms errors=0
large.jpg      resize    p50=  38.02ms p99= 135.23ms errors=0
large.jpg      webp      p50=  43.88ms p99= 426.91ms errors=0
large.jpg      avif      p50=  41.72ms p99= 301.58ms errors=0
medium.jpg     direct    p50=   0.54ms p99=   2.23ms errors=0
medium.jpg     raw       p50=  16.55ms p99=  87.56ms errors=0
medium.jpg     reencode  p50=  37.98ms p99=  59.72ms errors=0
medium.jpg     resize    p50=  17.79ms p99=  26.53ms errors=0
medium.jpg     webp      p50=  20.60ms p99=  81.32ms errors=0
medium.jpg     avif      p50=  20.80ms p99=  39.26ms errors=0
medium.png     direct    p50=   2.14ms p99=   5.20ms errors=0
medium.png     raw       p50= 119.08ms p99= 200.08ms errors=0
medium.png     reencode  p50= 261.43ms p99= 840.78ms errors=0
medium.png     resize    p50=  49.94ms p99= 645.00ms errors=0
medium.png     webp      p50=  53.51ms p99= 109.18ms errors=0
medium.png     avif      p50=  49.46ms p99= 629.24ms errors=0
small.jpg      direct    p50=   0.34ms p99=   2.69ms errors=0
small.jpg      raw       p50=   3.57ms p99=  20.33ms errors=0
small.jpg      reencode  p50=  12.12ms p99=  21.65ms errors=0
small.jpg      resize    p50=  14.44ms p99=  21.27ms errors=0
small.jpg      webp      p50=  18.31ms p99=  27.19ms errors=0
small.jpg      avif      p50=  17.23ms p99=  23.26ms errors=0

Latency (total time incl. body read), 500 requests, concurrency 8

       image  scenario    bytes  ttfb p50     p50     p90     p99      max    rps       p50 vs direct  errors
   large.jpg    direct    2.4MB      0.40    1.59    2.74    4.05     4.65   4605                   -       0
   large.jpg       raw    2.4MB      6.22   95.25  118.70  141.16   172.30     82    +93.66ms (60.0x)       0
   large.jpg  reencode    1.9MB    167.06  235.95  326.23  638.50   995.84     32  +234.36ms (148.7x)       0
   large.jpg    resize    3.9KB     37.69   38.02   58.41  135.23  3075.50    146    +36.44ms (24.0x)       0
   large.jpg      webp    2.0KB     43.83   43.88   62.68  426.91  2084.35    134    +42.29ms (27.6x)       0
   large.jpg      avif    2.9KB     41.69   41.72   67.66  301.58  2721.99     99    +40.14ms (26.3x)       0
  medium.jpg    direct  423.7KB      0.34    0.54    1.01    2.23     3.75  12613                   -       0
  medium.jpg       raw  423.7KB      4.38   16.55   22.05   87.56    92.61    449    +16.01ms (30.8x)       0
  medium.jpg  reencode  346.5KB     32.82   37.98   46.93   59.72    76.79    207    +37.45ms (70.7x)       0
  medium.jpg    resize    3.5KB     17.77   17.79   22.27   26.53    32.43    437    +17.25ms (33.1x)       0
  medium.jpg      webp    1.8KB     20.54   20.60   26.67   81.32   218.82    359    +20.06ms (38.4x)       0
  medium.jpg      avif    2.7KB     20.76   20.80   27.42   39.26   422.87    357    +20.26ms (38.7x)       0
  medium.png    direct    3.2MB      0.59    2.14    3.48    5.20     5.82   3419                   -       0
  medium.png       raw    3.2MB      6.18  119.08  136.49  200.08   224.29     67   +116.94ms (55.6x)       0
  medium.png  reencode    3.0MB    172.93  261.43  328.86  840.78  1992.39     28  +259.29ms (122.2x)       0
  medium.png    resize   60.4KB     49.81   49.94   76.07  645.00  2489.14     83    +47.80ms (23.3x)       0
  medium.png      webp    1.8KB     53.44   53.51   70.12  109.18   889.94    134    +51.37ms (25.0x)       0
  medium.png      avif    2.6KB     49.44   49.46   71.58  629.24  1526.72    103    +47.32ms (23.1x)       0
   small.jpg    direct   65.7KB      0.28    0.34    0.99    2.69     3.35  15692                   -       0
   small.jpg       raw   65.7KB      2.97    3.57    7.92   20.33    26.34   1743     +3.24ms (10.6x)       0
   small.jpg  reencode   54.4KB     11.95   12.12   16.52   21.65    23.04    633    +11.78ms (35.9x)       0
   small.jpg    resize    6.8KB     14.39   14.44   17.64   21.27    22.58    549    +14.11ms (42.9x)       0
   small.jpg      webp    3.3KB     18.28   18.31   22.30   27.19    31.81    431    +17.98ms (54.3x)       0
   small.jpg      avif    4.7KB     17.16   17.23   20.70   23.26    32.45    462    +16.89ms (51.1x)       0

All times in milliseconds.
```