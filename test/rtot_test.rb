require_relative 'test_helper'

describe 'rtot server' do
  before do
    delete(path: '/all')
  end

  it 'is pingable without auth' do
    get_request(path: '/', port: port).code.must_equal('200')
  end

  it 'requires auth to create jobs' do
    post_request(
      path: '/', body: 'echo huh',
      port: port
    ).code.must_equal('401')
  end

  it 'creates jobs' do
    post(path: '/', body: 'echo foo')[:res].code.must_equal('201')
  end

  it 'returns created job' do
    response = post(path: '/', body: 'echo huh')[:json]
    response['jobs'].wont_equal([])
    jobs = response['jobs']
    jobs.length.must_equal(1)
    jobs.first.keys.must_include('create')
  end

  it 'requires auth to get jobs' do
    response = post(path: '/', body: 'echo hurm')[:json]
    get_request(
      path: response['jobs'].first['href'],
      port: port
    ).code.must_equal('401')
  end

  it 'returns individual jobs' do
    response = post(path: '/', body: 'echo hurm ; sleep 1')[:json]
    get(path: response['jobs'].first['href'])[:res].code.must_equal('202')
  end

  it 'requires auth to get all jobs' do
    2.times { |n| post(path: '/', body: "echo #{n} time") }
    get_request(path: '/all', port: port).code.must_equal('401')
  end

  it 'returns all jobs' do
    2.times { |n| post(path: '/', body: "echo #{n} time") }
    get(path: '/all')[:json]['jobs'].length.must_be(:>, 1)
  end

  it 'allows for getting all by state' do
    post(path: '/', body: 'echo fast-ish')
    post(path: '/', body: 'echo slowwwww ; sleep 5')
    sleep 0.1
    get(path: '/all/running')[:json]['jobs'].length.must_equal(1)
  end

  it 'allows for deleting by state' do
    post(path: '/', body: 'echo fast-ish')
    post(path: '/', body: 'echo slowwwww ; sleep 5')
    sleep 0.1
    delete(path: '/all/complete')
    get(path: '/all')[:json]['jobs'].length.must_equal(1)
  end

  it 'includes "exit" as a string' do
    job = post(path: '/', body: 'exit 1')[:json]['jobs'].first
    sleep 0.1
    exit_str = get(path: job['href'])[:json]['jobs'].first['exit']
    exit_str.must_equal('exit status 1')
  end

  it 'omits "exit" when exit code is 0' do
    job = post(path: '/', body: 'exit 0')[:json]['jobs'].first
    sleep 0.1
    get(path: job['href'])[:json]['jobs'].first.wont_include('exit')
  end

  it 'omits "err" when empty' do
    job = post(path: '/', body: 'exit 0')[:json]['jobs'].first
    sleep 0.1
    get(path: job['href'])[:json]['jobs'].first.wont_include('err')
  end

  it 'omits "out" when empty' do
    job = post(path: '/', body: 'exit 0')[:json]['jobs'].first
    sleep 0.1
    get(path: job['href'])[:json]['jobs'].first.wont_include('out')
  end
end
