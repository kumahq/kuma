import{a as I,A as K}from"./AccordionList-BjuAwtfg.js";import{C as N}from"./CodeBlock-hG5z8uUD.js";import{d as T,a as r,o as t,b as _,w as e,e as o,c as n,F as d,G as h,a7 as b,f as a,t as f,m as l,q as g,p as z,v as M,x as j,_ as F}from"./index-BRR4OZXP.js";import{P as q}from"./PolicyTypeTag-LwaRrFLt.js";import{R as E}from"./RuleMatchers-CVXY9eF9.js";import{t as G}from"./toYaml-DB9FPXFY.js";const B=i=>(M("data-v-9dc2af64"),i=i(),j(),i),H={key:0,class:"rules"},Y=B(()=>l("h3",null,"Rules",-1)),J={class:"mt-4"},Q={class:"stack-with-borders"},U=B(()=>l("dt",null,`
                                    Config
                                  `,-1)),W={class:"mt-2"},X=T({__name:"ConnectionOutboundSummaryOverviewView",props:{data:{},dataplaneOverview:{}},setup(i){const V=i;return(Z,ee)=>{const P=r("KBadge"),R=r("RouterLink"),D=r("DataSource"),O=r("KCard"),S=r("DataCollection"),x=r("DataLoader"),L=r("AppView"),$=r("RouteView");return t(),_($,{params:{mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-overview-view"},{default:e(({t:A,route:y})=>[o(L,null,{default:e(()=>[(t(!0),n(d,null,h([y.params.connection.replace(/-([a-f0-9]){16}$/,"")],v=>(t(),n("div",{key:v,class:"stack-with-borders"},[o(b,{layout:"horizontal"},{title:e(()=>[a(`
              Protocol
            `)]),body:e(()=>[o(P,{appearance:"info"},{default:e(()=>[a(f(A(`http.api.value.${["grpc","http","tcp"].find(w=>typeof V.data[w]<"u")}`)),1)]),_:2},1024)]),_:2},1024),a(),V.data?(t(),n("div",H,[Y,a(),o(x,{src:`/meshes/${y.params.mesh}/rules/for/${y.params.dataPlane}`},{default:e(({data:w})=>[o(S,{predicate:p=>p.ruleType==="to"&&!["MeshHTTPRoute","MeshTCPRoute"].includes(p.type)&&p.matchers.every(s=>s.key==="kuma.io/service"&&(s.not?s.value!==v:s.value===v)),items:w.rules},{default:e(({items:p})=>[l("div",J,[o(I,{"initially-open":0,"multiple-open":"",class:"stack"},{default:e(()=>[(t(!0),n(d,null,h(Object.groupBy(p,s=>s.type),(s,k)=>(t(),_(O,{key:k},{default:e(()=>[o(K,null,{"accordion-header":e(()=>[o(q,{"policy-type":k},{default:e(()=>[a(f(k)+" ("+f(s.length)+`)
                            `,1)]),_:2},1032,["policy-type"])]),"accordion-content":e(()=>[l("div",Q,[(t(!0),n(d,null,h(s,u=>(t(),n(d,{key:u},[u.matchers.length>0?(t(),_(b,{key:0,layout:"horizontal"},{title:e(()=>[a(`
                                    To
                                  `)]),body:e(()=>[l("p",null,[o(E,{items:u.matchers},null,8,["items"])])]),_:2},1024)):g("",!0),a(),u.origins.length>0?(t(),_(b,{key:1,layout:"horizontal"},{title:e(()=>[a(`
                                    Origin Policies
                                  `)]),body:e(()=>[o(D,{src:"/policy-types"},{default:e(({data:C})=>[(t(!0),n(d,null,h([Object.groupBy((C==null?void 0:C.policies)??[],m=>m.name)],m=>(t(),n("ul",{key:m},[(t(!0),n(d,null,h(u.origins,c=>(t(),n("li",{key:`${c.mesh}-${c.name}`},[m[c.type]?(t(),_(R,{key:0,to:{name:"policy-detail-view",params:{mesh:c.mesh,policyPath:m[c.type][0].path,policy:c.name}}},{default:e(()=>[a(f(c.name),1)]),_:2},1032,["to"])):(t(),n(d,{key:1},[a(f(c.name),1)],64))]))),128))]))),128))]),_:2},1024)]),_:2},1024)):g("",!0),a(),l("div",null,[U,a(),l("dd",W,[l("div",null,[o(N,{code:z(G)(u.raw),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"])]),_:2},1032,["src"])])):g("",!0)]))),128))]),_:2},1024)]),_:1})}}}),re=F(X,[["__scopeId","data-v-9dc2af64"]]);export{re as default};
