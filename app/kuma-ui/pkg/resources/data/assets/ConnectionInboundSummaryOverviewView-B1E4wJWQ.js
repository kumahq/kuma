import{d as T,e as s,o as n,m as y,w as e,a,k as l,P as u,b as o,t as p,$ as B,p as b,c as r,l as V,a4 as j,H as c,J as v,a2 as F}from"./index-CjjKwNo4.js";import{a as I,A as K}from"./AccordionList-DA9rbXnW.js";import{P as M}from"./PolicyTypeTag-yDq7nHRi.js";import{T as E}from"./TagList-pzFi9naC.js";import{R as H}from"./RuleMatchers-C94XW3d9.js";const J={class:"stack-with-borders"},Y={key:0,class:"mt-6"},q={class:"mt-4"},G={class:"stack-with-borders"},Q={class:"mt-2"},at=T({__name:"ConnectionInboundSummaryOverviewView",props:{data:{}},setup(x){const d=x;return(U,t)=>{const P=s("XBadge"),D=s("XAction"),$=s("DataSource"),z=s("XCodeBlock"),h=s("KCard"),L=s("DataCollection"),N=s("DataLoader"),R=s("AppView"),X=s("RouteView");return n(),y(X,{params:{mesh:"",dataPlane:"",connection:""},name:"connection-inbound-summary-overview-view"},{default:e(({t:k,route:w,uri:S})=>[a(R,null,{default:e(()=>[l("div",J,[a(u,{layout:"horizontal"},{title:e(()=>t[0]||(t[0]=[o(`
            Tags
          `)])),body:e(()=>[a(E,{tags:d.data.tags,alignment:"right"},null,8,["tags"])]),_:1}),t[9]||(t[9]=o()),a(u,{layout:"horizontal"},{title:e(()=>[o(p(k("http.api.property.state")),1)]),body:e(()=>[a(P,{appearance:d.data.state==="Ready"?"success":"danger"},{default:e(()=>[o(p(k(`http.api.value.${d.data.state}`)),1)]),_:2},1032,["appearance"])]),_:2},1024),t[10]||(t[10]=o()),a(u,{layout:"horizontal"},{title:e(()=>t[3]||(t[3]=[o(`
            Protocol
          `)])),body:e(()=>[a(P,{appearance:"info"},{default:e(()=>[o(p(k(`http.api.value.${d.data.protocol}`)),1)]),_:2},1024)]),_:2},1024),t[11]||(t[11]=o()),a(u,{layout:"horizontal"},{title:e(()=>t[5]||(t[5]=[o(`
            Address
          `)])),body:e(()=>[a(B,{text:`${d.data.addressPort}`},null,8,["text"])]),_:1}),t[12]||(t[12]=o()),d.data.serviceAddressPort.length>0?(n(),y(u,{key:0,layout:"horizontal"},{title:e(()=>t[7]||(t[7]=[o(`
            Service Address
          `)])),body:e(()=>[a(B,{text:`${d.data.serviceAddressPort}`},null,8,["text"])]),_:1})):b("",!0)]),t[24]||(t[24]=o()),d.data?(n(),r("div",Y,[t[22]||(t[22]=l("h3",null,"Rules",-1)),t[23]||(t[23]=o()),a(N,{src:S(V(j),"/meshes/:mesh/rules/for/:dataplane",{mesh:w.params.mesh,dataplane:w.params.dataPlane})},{default:e(({data:O})=>[a(L,{predicate:_=>_.ruleType==="from"&&Number(_.inbound.port)===Number(w.params.connection.split("_")[1]),items:O.rules},{default:e(({items:_})=>[l("div",q,[a(I,{"initially-open":0,"multiple-open":"",class:"stack"},{default:e(()=>[(n(!0),r(c,null,v(Object.groupBy(_,g=>g.type),(g,C)=>(n(),y(h,{key:C},{default:e(()=>[a(K,null,{"accordion-header":e(()=>[a(M,{"policy-type":C},{default:e(()=>[o(p(C)+" ("+p(g.length)+`)
                        `,1)]),_:2},1032,["policy-type"])]),"accordion-content":e(()=>[l("div",G,[(n(!0),r(c,null,v(g,m=>(n(),r(c,{key:m},[m.matchers.length>0?(n(),y(u,{key:0,layout:"horizontal"},{title:e(()=>t[13]||(t[13]=[o(`
                                From
                              `)])),body:e(()=>[l("p",null,[a(H,{items:m.matchers},null,8,["items"])])]),_:2},1024)):b("",!0),t[19]||(t[19]=o()),m.origins.length>0?(n(),y(u,{key:1,layout:"horizontal"},{title:e(()=>t[15]||(t[15]=[o(`
                                Origin Policies
                              `)])),body:e(()=>[a($,{src:"/policy-types"},{default:e(({data:A})=>[(n(!0),r(c,null,v([Object.groupBy((A==null?void 0:A.policies)??[],f=>f.name)],f=>(n(),r("ul",{key:f},[(n(!0),r(c,null,v(m.origins,i=>(n(),r("li",{key:`${i.mesh}-${i.name}`},[f[i.type]?(n(),y(D,{key:0,to:{name:"policy-detail-view",params:{mesh:i.mesh,policyPath:f[i.type][0].path,policy:i.name}}},{default:e(()=>[o(p(i.name),1)]),_:2},1032,["to"])):(n(),r(c,{key:1},[o(p(i.name),1)],64))]))),128))]))),128))]),_:2},1024)]),_:2},1024)):b("",!0),t[20]||(t[20]=o()),l("div",null,[t[17]||(t[17]=l("dt",null,`
                                Config
                              `,-1)),t[18]||(t[18]=o()),l("dd",Q,[l("div",null,[a(z,{code:V(F).stringify(m.raw),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):b("",!0)]),_:2},1024)]),_:1})}}});export{at as default};
