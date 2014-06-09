//
//  FEAppDelegate.m
//  FintExample
//
//  Copyright (c) 2014 Soichiro Kashima. All rights reserved.
//
//  fint - A lightweight source code check tool.
//  https://github.com/ksoichiro/fint
//

#import "FEAppDelegate.h"

@implementation FEAppDelegate

// Test method for violation detection.
// Detect following format as violation:
//   if( --> NG
//   }else{  --> NG
//   if ((  --> OK
//   if ( ( --> NG
//   for( --> NG
//   ){ --> NG
//   ) ) { --> NG
//   alloc]init] --> NG
//   @"" ; --> NG
//   a=b --> NG
- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions
{
    self.window=[[UIWindow alloc]initWithFrame:[[UIScreen mainScreen] bounds]];
	// Override point for customization after application launch.
    self.window.backgroundColor = [UIColor whiteColor];  
    [self.window makeKeyAndVisible];
    if(floor(NSFoundationVersionNumber_iOS_6_1) < NSFoundationVersionNumber){
        //Do something
    }else if( YES ){
        // Do something
    }else{
        // Do something
        NSString *msg = [self doSomething:0 withOptions:@[@"foo",@"bar"]];
        NSLog(@"a / b = %@", msg) ;
    }
    NSString *url = @"https://github.com/ksoichiro/fint";
    // a,b,c
    int x = 1;
    if (x == 2) {
    }
    x /= 1;
    x *= 2;
    x += 3;
    x -= 4;
    x %= 2;
    x/= 2;
    x /=8;
    x/=1;
    x*=2;
    x+=3;
    x-=4;
    x%=2;
    if (x != 2) {
    }
    if (x <= 2) {
    }
    if (x >= 2) {
    }
    if (x!=2) {
    }
    if (x<=2) {
    }
    if (x>=2) {
    }
    if ((x * 2) < 10) {
    }
    BOOL a = NO, b = YES;
    if (a && b) {
    }
    if (a || b) {
    }
    if (a&&b) {
    }
    if (a||b) {
    }
    a &= YES;
    a |= YES;
    a&=YES;
    a|=YES;
    for(int i=0;i<=10;i++){
        ;
        ; // Empty ; Dummy
    }
    for (int i= 0;i<= 10;i++) {
    } // a=b
    switch(x) {
        case 1:
            NSLog(@"1");
            break;
        default:
            break;
    }
    @"   if("; // --> NG
    @"   }else{"; // --> NG
    @"   if (("; // --> OK
    @"   if ( ("; // --> NG
    @"   for("; // --> NG
    @"   ){"; // --> NG
    @"   ) ) {"; // --> NG
    @"   alloc]init]"; // --> NG
    @"   a=b"; // --> NG
    @"   a,b,c"; // --> NG
    return YES;
}

- (void)applicationWillResignActive:(UIApplication *)application
{
    // Sent when the application is about to move from active to inactive state. This can occur for certain types of temporary interruptions (such as an incoming phone call or SMS message) or when the user quits the application and it begins the transition to the background state.
    // Use this method to pause ongoing tasks, disable timers, and throttle down OpenGL ES frame rates. Games should use this method to pause the game.
}

- (void)applicationDidEnterBackground:(UIApplication *)application
{
    // Use this method to release shared resources, save user data, invalidate timers, and store enough application state information to restore your application to its current state in case it is terminated later. 
    // If your application supports background execution, this method is called instead of applicationWillTerminate: when the user quits.
}

- (void)applicationWillEnterForeground:(UIApplication *)application
{
    // Called as part of the transition from the background to the inactive state; here you can undo many of the changes made on entering the background.
}

- (void)applicationDidBecomeActive:(UIApplication *)application
{
    // Restart any tasks that were paused (or not yet started) while the application was inactive. If the application was previously in the background, optionally refresh the user interface.
}

- (void)applicationWillTerminate:(UIApplication *)application
{
    // Called when the application is about to terminate. Save data if appropriate. See also applicationDidEnterBackground:.
}

- (NSString *)doSomething:(int)type withOptions:(NSArray *)options
{
    return @"Hello, World!";
}

@end
