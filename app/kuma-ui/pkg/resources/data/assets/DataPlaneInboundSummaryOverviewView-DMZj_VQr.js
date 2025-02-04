import{d as M,r as l,o as n,q as p,w as e,b as a,m as r,U as u,e as o,t as m,s as g,c as d,p as P,a0 as T,M as f,N as k,a1 as j}from"./index-DhvxqWBA.js";import{A as F,a as I}from"./AccordionList-C697zAlf.js";import{P as q}from"./PolicyTypeTag-ByoCWwaQ.js";import{T as E}from"./TagList-D35hhA-j.js";import{R as U}from"./RuleMatchers-CXWLkz9O.js";const Y={class:"stack-with-borders"},G={key:0,class:"mt-6"},H={class:"mt-4"},J={class:"stack-with-borders"},K={class:"mt-2"},at=M({__name:"DataPlaneInboundSummaryOverviewView",props:{data:{},routeName:{}},setup(x){const s=x;return(Q,t)=>{const N=l("XBadge"),v=l("XCopyButton"),X=l("XAction"),D=l("DataSource"),V=l("XCodeBlock"),z=l("XCard"),$=l("DataCollection"),L=l("DataLoader"),R=l("AppView"),S=l("RouteView");return n(),p(S,{params:{mesh:"",dataPlane:"",connection:""},name:s.routeName},{default:e(({t:C,route:w,uri:h})=>[a(R,null,{default:e(()=>[r("div",Y,[a(u,{layout:"horizontal"},{title:e(()=>t[0]||(t[0]=[o(`
            Tags
          `)])),body:e(()=>[a(E,{tags:s.data.tags,alignment:"right"},null,8,["tags"])]),_:1}),t[11]||(t[11]=o()),a(u,{layout:"horizontal"},{title:e(()=>[o(m(C("http.api.property.state")),1)]),body:e(()=>[a(N,{appearance:s.data.state==="Ready"?"success":"danger"},{default:e(()=>[o(m(C(`http.api.value.${s.data.state}`)),1)]),_:2},1032,["appearance"])]),_:2},1024),t[12]||(t[12]=o()),a(u,{layout:"horizontal"},{title:e(()=>t[3]||(t[3]=[o(`
            Protocol
          `)])),body:e(()=>[a(N,{appearance:"info"},{default:e(()=>[o(m(C(`http.api.value.${s.data.protocol}`)),1)]),_:2},1024)]),_:2},1024),t[13]||(t[13]=o()),a(u,{layout:"horizontal"},{title:e(()=>t[5]||(t[5]=[o(`
            Address
          `)])),body:e(()=>[a(v,{text:`${s.data.addressPort}`},null,8,["text"])]),_:1}),t[14]||(t[14]=o()),s.data.serviceAddressPort.length>0?(n(),p(u,{key:0,layout:"horizontal"},{title:e(()=>t[7]||(t[7]=[o(`
            Service Address
          `)])),body:e(()=>[a(v,{text:`${s.data.serviceAddressPort}`},null,8,["text"])]),_:1})):g("",!0),t[15]||(t[15]=o()),s.data.portName.length>0?(n(),p(u,{key:1,layout:"horizontal"},{title:e(()=>t[9]||(t[9]=[o(`
            Name
          `)])),body:e(()=>[a(v,{text:`${s.data.portName}`},null,8,["text"])]),_:1})):g("",!0)]),t[27]||(t[27]=o()),s.data?(n(),d("div",G,[t[25]||(t[25]=r("h3",null,"Rules",-1)),t[26]||(t[26]=o()),a(L,{src:h(P(T),"/meshes/:mesh/rules/for/:dataplane",{mesh:w.params.mesh,dataplane:w.params.dataPlane})},{default:e(({data:O})=>[a($,{predicate:_=>_.ruleType==="from"&&Number(_.inbound.port)===Number(w.params.connection.split("_")[1]),items:O.rules},{default:e(({items:_})=>[r("div",H,[a(F,{"initially-open":0,"multiple-open":"",class:"stack"},{default:e(()=>[(n(!0),d(f,null,k(Object.groupBy(_,b=>b.type),(b,A)=>(n(),p(z,{key:A},{default:e(()=>[a(I,null,{"accordion-header":e(()=>[a(q,{"policy-type":A},{default:e(()=>[o(m(A)+" ("+m(b.length)+`)
                        `,1)]),_:2},1032,["policy-type"])]),"accordion-content":e(()=>[r("div",J,[(n(!0),d(f,null,k(b,y=>(n(),d(f,{key:y},[y.matchers.length>0?(n(),p(u,{key:0,layout:"horizontal"},{title:e(()=>t[16]||(t[16]=[o(`
                                From
                              `)])),body:e(()=>[r("p",null,[a(U,{items:y.matchers},null,8,["items"])])]),_:2},1024)):g("",!0),t[22]||(t[22]=o()),y.origins.length>0?(n(),p(u,{key:1,layout:"horizontal"},{title:e(()=>t[18]||(t[18]=[o(`
                                Origin Policies
                              `)])),body:e(()=>[a(D,{src:"/policy-types"},{default:e(({data:B})=>[(n(!0),d(f,null,k([Object.groupBy((B==null?void 0:B.policies)??[],c=>c.name)],c=>(n(),d("ul",{key:c},[(n(!0),d(f,null,k(y.origins,i=>(n(),d("li",{key:`${i.mesh}-${i.name}`},[c[i.type]?(n(),p(X,{key:0,to:{name:"policy-detail-view",params:{mesh:i.mesh,policyPath:c[i.type][0].path,policy:i.name}}},{default:e(()=>[o(m(i.name),1)]),_:2},1032,["to"])):(n(),d(f,{key:1},[o(m(i.name),1)],64))]))),128))]))),128))]),_:2},1024)]),_:2},1024)):g("",!0),t[23]||(t[23]=o()),r("div",null,[t[20]||(t[20]=r("dt",null,`
                                Config
                              `,-1)),t[21]||(t[21]=o()),r("dd",K,[r("div",null,[a(V,{code:P(j).stringify(y.raw),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):g("",!0)]),_:2},1024)]),_:1},8,["name"])}}});export{at as default};
