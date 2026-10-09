# Kubling gRPC Java client

Generated messages, gRPC stubs and `Features` constants for the official Kubling
protocol. The supported Java release is declared in the POM. Server support is
discovered through capabilities; including a definition in this JAR does not enable it on a server.
`PARTIAL_RESULTS_V1` gates terminal completeness and structured warnings.
`WarningRole` distinguishes general diagnostics from partial-result causes;
`allowPartialResults` authorizes incomplete query data for one Execute call.

Coordinates: `com.kubling:kubling-grpc`. Official releases share the canonical
repository version; availability depends on a completed release to Maven Central.

```xml
<dependency>
  <groupId>com.kubling</groupId>
  <artifactId>kubling-grpc</artifactId>
  <version>1.2.0</version>
</dependency>
```

Applications supply a gRPC transport, for example `io.grpc:grpc-netty-shaded`
aligned with the gRPC BOM. The library does not select a transport or configure
authentication, retries or transaction recovery.

```java
import com.kubling.transport.grpc.QueryServiceGrpc;

var query = QueryServiceGrpc.newBlockingStub(channel);
```

Build from the repository checkout with GraalVM, Bash, Python 3 and network
access to the Buf generators and Maven repositories. The workflow uses Oracle
GraalVM 25 and validates the build JDK independently from the Java 21 artifact
compatibility target declared in the POM:

```sh
bash tools/check_java_build.sh
./mvnw --batch-mode --no-transfer-progress -f sdk-java/pom.xml clean verify
```

The build generates Java from the canonical `../proto` directory. Generated
sources stay out of Git. The output contains the binary JAR, sources JAR and
Javadoc JAR; the binary also includes the canonical protos, feature registry
and license under `META-INF`. Consumers need neither Buf nor protoc.

See [release instructions](../docs/releases.md).
