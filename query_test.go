package xmlquery

import (
	"fmt"
	"strings"
	"testing"
)

// https://msdn.microsoft.com/en-us/library/ms762271(v=vs.85).aspx
const xmlDoc = `
    <?xml version="1.0"?>
<catalog>
   <!-- book list-->
   <book id="bk101">
      <author>Gambardella, Matthew</author>
      <title>XML Developer's Guide</title>
      <genre>Computer</genre>
      <price>44.95</price>
      <publish_date>2000-10-01</publish_date>
      <description>An in-depth look at creating applications
      with XML.</description>
   </book>
   <book id="bk102">
      <author>Ralls, Kim</author>
      <title>Midnight Rain</title>
      <genre>Fantasy</genre>
      <price>5.95</price>
      <publish_date>2000-12-16</publish_date>
      <description>A former architect battles corporate zombies,
      an evil sorceress, and her own childhood to become queen
      of the world.</description>
   </book>
   <book id="bk103">
      <author>Corets, Eva</author>
      <title>Maeve Ascendant</title>
      <genre>Fantasy</genre>
      <price>5.95</price>
      <publish_date>2000-11-17</publish_date>
      <description>After the collapse of a nanotechnology
      society in England, the young survivors lay the
      foundation for a new society.</description>
   </book>
</catalog>`

var doc = loadXML(xmlDoc)

func TestXPath(t *testing.T) {
	if list := Find(doc, "//book"); len(list) != 3 {
		t.Fatal("count(//book) != 3")
	}
	if node := FindOne(doc, "//book[@id='bk101']"); node == nil {
		t.Fatal("//book[@id='bk101] is not found")
	}
	if node := FindOne(doc, "//book[price>=44.95]"); node == nil {
		t.Fatal("//book/price>=44.95 is not found")
	}
	if list := Find(doc, "//book[genre='Fantasy']"); len(list) != 2 {
		t.Fatal("//book[genre='Fantasy'] items count is not equal 2")
	}
	var c int
	FindEach(doc, "//book", func(i int, n *Node) {
		c++
	})
	l := len(Find(doc, "//book"))
	if c != l {
		t.Fatal("count(//book) != 3")
	}
	c = 0
	FindEachWithBreak(doc, "//book", func(i int, n *Node) bool {
		if c == l-1 {
			return false
		}
		c++
		return true
	})
	if c != l-1 {
		t.Fatal("FindEachWithBreak failed to stop.")
	}
	node := FindOne(doc, "//book[1]")
	if node.SelectAttr("id") != "bk101" {
		t.Fatal("//book[1]/@id != bk101")
	}
}

func TestXPathCdUp(t *testing.T) {
	doc := loadXML(`<a><b attr="1"/></a>`)
	node := FindOne(doc, "/a/b/@attr/..")
	t.Logf("node = %#v", node)
	if node == nil || node.Data != "b" {
		t.Fatal("//b/@id/.. != <b/>")
	}
}

func TestInvalidXPathExpression(t *testing.T) {
	doc := &Node{}
	_, err := QueryAll(doc, "//a[@a==1]")
	if err == nil {
		t.Fatal("expected a parsed error but nil")
	}
	_, err = Query(doc, "//a[@a==1]")
	if err == nil {
		t.Fatal("expected a parsed error but nil")
	}
}

func TestNavigator(t *testing.T) {
	nav := &NodeNavigator{curr: doc, root: doc, attr: -1}
	nav.MoveToChild() // New Line
	nav.MoveToNext()  // catalog
	if nav.curr.Data != "catalog" {
		t.Fatal("current node name != `catalog`")
	}
	nav.MoveToChild() // New Line
	nav.MoveToNext()  // comment node
	if nav.curr.Type != CommentNode {
		t.Fatal("node type not CommentNode")
	}
	nav.Value()
	nav.MoveToNext() //book
	nav.MoveToChild()
	nav.MoveToNext() // book/author
	if nav.LocalName() != "author" {
		t.Fatalf("node error")
	}
	nav.MoveToParent() // book
	nav.MoveToNext()   // next book
	if nav.curr.SelectAttr("id") != "bk102" {
		t.Fatal("node error")
	}
}

func TestAttributesNamespaces(t *testing.T) {
	doc := loadXML(`
		<root xmlns="ns://root" xmlns:nested="ns://nested" xmlns:other="ns://other">
			<tag id="1" attr="v"></tag>
			<tag id="2" nested:attr="v"></tag>
			<nested:tag id="3" nested:attr="v"></nested:tag>
			<nested:tag id="4" other:attr="v"></nested:tag>
			<nested:tag id="5" attr="v"></nested:tag>
		</root>
	`)
	results := Find(doc, "//*[@*[namespace-uri()='ns://nested' and local-name()='attr']]")
	parsed := make([]string, 0, 5)
	for _, tag := range results {
		parsed = append(parsed, tag.SelectAttr("id"))
	}
	got := fmt.Sprintf("%v", parsed)

	// unsure if 5 should be selected here
	if got != "[2 3]" {
		t.Fatalf("Expected tags [2 3], got %v", got)
	}
}

func TestElementsNamespaceWithoutPrefix(t *testing.T) {
	doc := loadXML(`
	<Objects xmlns="http://example.com/schema/2007/someschema">
		<Object>
			<SourceInformation>
				<Provider Name="Object Provider" Guid="{3CB2A168-FE19-4A4E-BDAD-DCF422F13473}"/>
				<ObjectType>4663</ObjectType>
			</SourceInformation>
			<ObjectData>
				<Data Name="ObjectId">ObjectA</Data>
				<Data Name="ObjectName">ROFL</Data>
			</ObjectData>
		</Object>
		<Object>
			<SourceInformation>
				<Provider Name="Object Provider" Guid="{3CB2A168-FE19-4A4E-BDAD-DCF422F13473}"/>
				<ObjectType>4663</ObjectType>
			</SourceInformation>
			<ObjectData>
				<Data Name="ObjectId">ObjectB</Data>
				<Data Name="ObjectName">TASTIC</Data>
			</ObjectData>
		</Object>
		<Object>
			<SourceInformation>
				<Provider Name="Object Provider" Guid="{3CB2A168-FE19-4A4E-BDAD-DCF422F13473}"/>
				<ObjectType>4663</ObjectType>
			</SourceInformation>
			<ObjectData>
				<Data Name="ObjectId">ObjectC</Data>
				<Data Name="ObjectName">LMAO</Data>
			</ObjectData>
		</Object>
	</Objects>
	`)
	results := Find(doc, "//Objects/*[namespace-uri()=\"http://example.com/schema/2007/someschema\" and local-name()=\"Object\"]/ObjectData/Data[@Name=\"ObjectId\"]")
	parsed := make([]string, 0, 5)
	for _, tag := range results {
		parsed = append(parsed, tag.InnerText() /*tag.SelectAttr("id")*/)
	}
	got := fmt.Sprintf("%v", parsed)

	if got != "[ObjectA ObjectB ObjectC]" {
		t.Fatalf("Expected data [ObjectA ObjectB ObjectC], got %v", got)
	}
}

func loadXML(s string) *Node {
	node, err := Parse(strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	return node
}

func TestMissingTextNodes(t *testing.T) {
	doc := loadXML(`
        <?xml version="1.0" encoding="utf-8"?>
        <corpus><p>Lorem <a>ipsum</a> dolor</p></corpus>
    `)
	results := Find(doc, "//text()")
	if len(results) != 3 {
		t.Fatalf("Expected text nodes 3, got %d", len(results))
	}
}
