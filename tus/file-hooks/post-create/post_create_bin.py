import sys, getopt
import json

def main(argv):
  metadata = ''
  opts, args = getopt.getopt(argv, "hm:", ['metadata='])

  for opt, arg in opts:
    if opt == '-h':
      print('post_create_bin.py -m <inputfile>')
      sys.exit()
    elif opt in ("-m", "--metadata"):
      metadata = arg
  

if __name__ == "__main__":
  main(sys.argv[1:])
