import{d as X,r as p,o as e,m as d,w as t,b as c,c as o,L as i,M as _,Z as M,e as a,t as h,l as B,aC as q,k as l,p as b,a7 as $,H as E,J as Y,q as Z}from"./index-B3PYX6oN.js";import{a as G,A as Q}from"./AccordionList-DwM5irqL.js";import{C as A}from"./CodeBlock-Cf09PDeH.js";import{P as L}from"./PolicyTypeTag-0xgJslFo.js";import{R as U}from"./RuleMatchers-DPmln6Db.js";const D=v=>(E("data-v-8cc89122"),v=v(),Y(),v),W={key:0,class:"rules"},ee=D(()=>l("h3",null,"Rules",-1)),te={class:"stack-with-borders mt-4"},ae={class:"stack-with-borders mt-4"},oe={class:"mt-4"},ne={class:"stack-with-borders"},se=D(()=>l("dt",null,`
                                          Config
                                        `,-1)),re={class:"mt-2"},ce=X({__name:"ConnectionOutboundSummaryOverviewView",props:{data:{},dataplaneOverview:{}},setup(v){const k=v,S=(w,f)=>w.$resourceMeta.name===f.name&&w.$resourceMeta.namespace===f.namespace&&w.$resourceMeta.zone===f.zone&&(f.resourceSectionName===""||w.$resourceMeta.port===f.port);return(w,f)=>{const z=p("KBadge"),N=p("XAction"),O=p("DataCollection"),x=p("RouterLink"),I=p("KCard"),j=p("DataLoader"),K=p("DataSource"),F=p("AppView"),T=p("RouteView");return e(),d(T,{params:{mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-overview-view"},{default:t(({t:H,route:P,uri:J})=>[c(F,null,{default:t(()=>[(e(!0),o(i,null,_([P.params.connection.replace(/-([a-f0-9]){16}$/,"")],R=>(e(),o("div",{key:R,class:"stack-with-borders"},[c(M,{layout:"horizontal"},{title:t(()=>[a(`
              Protocol
            `)]),body:t(()=>[c(z,{appearance:"info"},{default:t(()=>[a(h(H(`http.api.value.${["grpc","http","tcp"].find(g=>typeof k.data[g]<"u")}`)),1)]),_:2},1024)]),_:2},1024),a(),k.data?(e(),o("div",W,[ee,a(),c(K,{src:"/policy-types"},{default:t(({data:g})=>[(e(!0),o(i,null,_([Object.groupBy((g==null?void 0:g.policies)??[],y=>y.name)],y=>(e(),d(j,{key:typeof y,src:J(B(q),"/meshes/:mesh/rules/for/:dataplane",{mesh:P.params.mesh,dataplane:P.params.dataPlane})},{default:t(({data:V})=>[k.data.$resourceMeta.type!==""?(e(),d(O,{key:0,predicate:u=>u.resourceMeta.type==="Mesh"||S(k.data,u),items:V.toResourceRules},{default:t(({items:u})=>[l("div",te,[(e(!0),o(i,null,_(Object.groupBy(u,n=>n.type),(n,m)=>(e(),o("div",{key:m},[c(L,{"policy-type":m},{default:t(()=>[a(h(m),1)]),_:2},1032,["policy-type"]),a(),l("div",ae,[(e(!0),o(i,null,_(n.length>1?n.filter(s=>S(k.data,s)):n,s=>(e(),o("div",{key:s},[s.origins.length>0?(e(),d(M,{key:0,layout:"horizontal"},{title:t(()=>[a(`
                                      Origin Policies
                                    `)]),body:t(()=>[c(O,{predicate:r=>typeof r.resourceMeta<"u",items:s.origins,empty:!1},{default:t(({items:r})=>[l("ul",null,[(e(!0),o(i,null,_(r,C=>(e(),o("li",{key:JSON.stringify(C)},[Object.keys(y).length>0?(e(),d(N,{key:0,to:{name:"policy-detail-view",params:{policyPath:y[m][0].path,mesh:C.resourceMeta.mesh,policy:C.resourceMeta.name}}},{default:t(()=>[a(h(C.resourceMeta.name),1)]),_:2},1032,["to"])):b("",!0)]))),128))])]),_:2},1032,["predicate","items"])]),_:2},1024)):b("",!0),a(),c(A,{class:"mt-2",code:B($).stringify(s.raw),language:"yaml","show-copy-button":!1},null,8,["code"])]))),128))])]))),128))])]),_:2},1032,["predicate","items"])):(e(),d(O,{key:1,predicate:u=>u.ruleType==="to"&&!["MeshHTTPRoute","MeshTCPRoute"].includes(u.type)&&u.matchers.every(n=>n.key==="kuma.io/service"&&(n.not?n.value!==R:n.value===R)),items:V.rules},{default:t(({items:u})=>[l("div",oe,[c(G,{"initially-open":0,"multiple-open":"",class:"stack"},{default:t(()=>[(e(!0),o(i,null,_(Object.groupBy(u,n=>n.type),(n,m)=>(e(),d(I,{key:m},{default:t(()=>[c(Q,null,{"accordion-header":t(()=>[c(L,{"policy-type":m},{default:t(()=>[a(h(m)+" ("+h(n.length)+`)
                                  `,1)]),_:2},1032,["policy-type"])]),"accordion-content":t(()=>[l("div",ne,[(e(!0),o(i,null,_(n,s=>(e(),o(i,{key:s},[s.matchers.length>0?(e(),d(M,{key:0,layout:"horizontal"},{title:t(()=>[a(`
                                          From
                                        `)]),body:t(()=>[l("p",null,[c(U,{items:s.matchers},null,8,["items"])])]),_:2},1024)):b("",!0),a(),s.origins.length>0?(e(),d(M,{key:1,layout:"horizontal"},{title:t(()=>[a(`
                                          Origin Policies
                                        `)]),body:t(()=>[l("ul",null,[(e(!0),o(i,null,_(s.origins,r=>(e(),o("li",{key:`${r.mesh}-${r.name}`},[y[r.type]?(e(),d(x,{key:0,to:{name:"policy-detail-view",params:{mesh:r.mesh,policyPath:y[r.type][0].path,policy:r.name}}},{default:t(()=>[a(h(r.name),1)]),_:2},1032,["to"])):(e(),o(i,{key:1},[a(h(r.name),1)],64))]))),128))])]),_:2},1024)):b("",!0),a(),l("div",null,[se,a(),l("dd",re,[l("div",null,[c(A,{code:B($).stringify(s.raw),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"]))]),_:2},1032,["src"]))),128))]),_:2},1024)])):b("",!0)]))),128))]),_:2},1024)]),_:1})}}}),me=Z(ce,[["__scopeId","data-v-8cc89122"]]);export{me as default};
