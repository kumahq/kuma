import{d as E,r as m,o as e,m as u,w as t,b as c,c as o,L as i,M as h,Z as M,e as a,t as _,p as w,l as R,aC as F,k as d,a7 as V,H,J,q as X}from"./index-BZS2OlAU.js";import{a as Z,A as q}from"./AccordionList-BmEd6nMz.js";import{C as $}from"./CodeBlock-CzUB6YWi.js";import{P as x}from"./PolicyTypeTag-BV-o7-1D.js";import{R as Y}from"./RuleMatchers-BcLI8LFB.js";const A=k=>(H("data-v-dbd888ff"),k=k(),J(),k),G={key:1,class:"rules"},Q=A(()=>d("h3",null,"Rules",-1)),U={class:"stack-with-borders mt-4"},W={class:"stack-with-borders mt-4"},ee={class:"mt-4"},te={class:"stack-with-borders"},ae=A(()=>d("dt",null,`
                                          Config
                                        `,-1)),oe={class:"mt-2"},ne=E({__name:"ConnectionOutboundSummaryOverviewView",props:{data:{},dataplaneOverview:{}},setup(k){const f=k;return(se,re)=>{const S=m("XAction"),P=m("KBadge"),g=m("DataCollection"),L=m("RouterLink"),D=m("KCard"),z=m("DataLoader"),I=m("DataSource"),N=m("AppView"),j=m("RouteView");return e(),u(j,{params:{mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-overview-view"},{default:t(({t:K,route:C,uri:T})=>[c(N,null,{default:t(()=>[(e(!0),o(i,null,h([C.params.connection.replace(/-([a-f0-9]){16}$/,"")],O=>(e(),o("div",{key:O,class:"stack-with-borders"},[f.data.$resourceMeta.type!==""?(e(),u(M,{key:0,layout:"horizontal"},{title:t(()=>[a(`
                Resource
              `)]),body:t(()=>[c(P,{appearance:"info","max-width":"auto"},{default:t(()=>[(e(!0),o(i,null,h([f.data.$resourceMeta],n=>(e(),u(S,{key:typeof n,to:{MeshService:{name:"mesh-service-detail-view",params:{mesh:n.mesh,service:n.name}},MeshExternalService:{name:"mesh-external-service-detail-view",params:{mesh:n.mesh,service:n.name}},MeshMultiZoneService:{name:"mesh-multizone-service-detail-view",params:{mesh:n.mesh,service:n.name}}}[n.type]},{default:t(()=>[a(_(n.type)+": "+_(n.name),1)]),_:2},1032,["to"]))),128))]),_:1})]),_:1})):w("",!0),a(),c(M,{layout:"horizontal"},{title:t(()=>[a(`
              Protocol
            `)]),body:t(()=>[c(P,{appearance:"info"},{default:t(()=>[a(_(K(`http.api.value.${["grpc","http","tcp"].find(n=>typeof f.data[n]<"u")}`)),1)]),_:2},1024)]),_:2},1024),a(),f.data?(e(),o("div",G,[Q,a(),c(I,{src:"/policy-types"},{default:t(({data:n})=>[(e(!0),o(i,null,h([Object.groupBy((n==null?void 0:n.policies)??[],v=>v.name)],v=>(e(),u(z,{key:typeof v,src:T(R(F),"/meshes/:mesh/rules/for/:dataplane",{mesh:C.params.mesh,dataplane:C.params.dataPlane})},{default:t(({data:B})=>[f.data.$resourceMeta.type!==""?(e(),u(g,{key:0,predicate:p=>p.resourceMeta.type==="Mesh"||f.data.$resourceMeta.name===p.resourceMeta.name,items:B.toResourceRules},{default:t(({items:p})=>[d("div",U,[(e(!0),o(i,null,h(Object.groupBy(p,s=>s.type),(s,y)=>(e(),o("div",{key:y},[c(x,{"policy-type":y},{default:t(()=>[a(_(y),1)]),_:2},1032,["policy-type"]),a(),d("div",W,[(e(!0),o(i,null,h(s.length>1?s.filter(r=>f.data.$resourceMeta.name===r.resourceMeta.name):s,r=>(e(),o("div",{key:r},[r.origins.length>0?(e(),u(M,{key:0,layout:"horizontal"},{title:t(()=>[a(`
                                      Origin Policies
                                    `)]),body:t(()=>[c(g,{predicate:l=>typeof l.resourceMeta<"u",items:r.origins,empty:!1},{default:t(({items:l})=>[d("ul",null,[(e(!0),o(i,null,h(l,b=>(e(),o("li",{key:JSON.stringify(b)},[Object.keys(v).length>0?(e(),u(S,{key:0,to:{name:"policy-detail-view",params:{policyPath:v[y][0].path,mesh:b.resourceMeta.mesh,policy:b.resourceMeta.name}}},{default:t(()=>[a(_(b.resourceMeta.name),1)]),_:2},1032,["to"])):w("",!0)]))),128))])]),_:2},1032,["predicate","items"])]),_:2},1024)):w("",!0),a(),c($,{class:"mt-2",code:R(V).stringify(r.raw),language:"yaml","show-copy-button":!1},null,8,["code"])]))),128))])]))),128))])]),_:2},1032,["predicate","items"])):(e(),u(g,{key:1,predicate:p=>p.ruleType==="to"&&!["MeshHTTPRoute","MeshTCPRoute"].includes(p.type)&&p.matchers.every(s=>s.key==="kuma.io/service"&&(s.not?s.value!==O:s.value===O)),items:B.rules},{default:t(({items:p})=>[d("div",ee,[c(Z,{"initially-open":0,"multiple-open":"",class:"stack"},{default:t(()=>[(e(!0),o(i,null,h(Object.groupBy(p,s=>s.type),(s,y)=>(e(),u(D,{key:y},{default:t(()=>[c(q,null,{"accordion-header":t(()=>[c(x,{"policy-type":y},{default:t(()=>[a(_(y)+" ("+_(s.length)+`)
                                  `,1)]),_:2},1032,["policy-type"])]),"accordion-content":t(()=>[d("div",te,[(e(!0),o(i,null,h(s,r=>(e(),o(i,{key:r},[r.matchers.length>0?(e(),u(M,{key:0,layout:"horizontal"},{title:t(()=>[a(`
                                          From
                                        `)]),body:t(()=>[d("p",null,[c(Y,{items:r.matchers},null,8,["items"])])]),_:2},1024)):w("",!0),a(),r.origins.length>0?(e(),u(M,{key:1,layout:"horizontal"},{title:t(()=>[a(`
                                          Origin Policies
                                        `)]),body:t(()=>[d("ul",null,[(e(!0),o(i,null,h(r.origins,l=>(e(),o("li",{key:`${l.mesh}-${l.name}`},[v[l.type]?(e(),u(L,{key:0,to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:v[l.type][0].path,policy:l.name}}},{default:t(()=>[a(_(l.name),1)]),_:2},1032,["to"])):(e(),o(i,{key:1},[a(_(l.name),1)],64))]))),128))])]),_:2},1024)):w("",!0),a(),d("div",null,[ae,a(),d("dd",oe,[d("div",null,[c($,{code:R(V).stringify(r.raw),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"]))]),_:2},1032,["src"]))),128))]),_:2},1024)])):w("",!0)]))),128))]),_:2},1024)]),_:1})}}}),pe=X(ne,[["__scopeId","data-v-dbd888ff"]]);export{pe as default};
