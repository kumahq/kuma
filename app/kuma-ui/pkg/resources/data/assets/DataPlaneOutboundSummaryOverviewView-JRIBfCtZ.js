import{d as q,r as m,o as t,q as p,w as o,b as u,c as n,M as c,N as f,U as M,e as a,t as h,m as d,p as P,a0 as E,s as g,a1 as $,_ as H}from"./index-BP47cGGe.js";import{A as J,a as U}from"./AccordionList-Li6Jw-zc.js";import{P as A}from"./PolicyTypeTag-akbPYmnj.js";import{R as Y}from"./RuleMatchers-B0v-dSkk.js";const G={key:0,class:"rules"},K={class:"stack-with-borders mt-4"},Q={class:"stack-with-borders mt-4"},W={class:"mt-4"},Z={class:"stack-with-borders"},ee={class:"mt-2"},te=q({__name:"DataPlaneOutboundSummaryOverviewView",props:{data:{},routeName:{}},setup(X){const k=X,V=(v,e)=>v.$resourceMeta.name===e.name&&v.$resourceMeta.namespace===e.namespace&&v.$resourceMeta.zone===e.zone&&(e.resourceSectionName===""||v.$resourceMeta.port===e.port);return(v,e)=>{const x=m("XBadge"),D=m("XAction"),C=m("DataCollection"),N=m("XCodeBlock"),S=m("XCard"),z=m("DataLoader"),L=m("DataSource"),j=m("AppView"),F=m("RouteView");return t(),p(F,{params:{mesh:"",proxy:"",connection:""},name:k.routeName},{default:o(({t:T,route:B,uri:I})=>[u(j,null,{default:o(()=>[(t(!0),n(c,null,f([B.params.connection.replace(/-([a-f0-9]){16}$/,"")],O=>(t(),n("div",{key:O,class:"stack-with-borders"},[u(M,{layout:"horizontal"},{title:o(()=>e[0]||(e[0]=[a(`
              Protocol
            `)])),body:o(()=>[u(x,{appearance:"info"},{default:o(()=>[a(h(T(`http.api.value.${["grpc","http","tcp"].find(w=>typeof k.data[w]<"u")}`)),1)]),_:2},1024)]),_:2},1024),e[17]||(e[17]=a()),k.data?(t(),n("div",G,[e[15]||(e[15]=d("h3",null,"Rules",-1)),e[16]||(e[16]=a()),u(L,{src:"/policy-types"},{default:o(({data:w})=>[(t(!0),n(c,null,f([Object.groupBy((w==null?void 0:w.policies)??[],_=>_.name)],_=>(t(),p(z,{key:typeof _,src:I(P(E),"/meshes/:mesh/rules/for/:dataplane",{mesh:B.params.mesh,dataplane:B.params.proxy})},{default:o(({data:R})=>[k.data.$resourceMeta.type!==""?(t(),p(C,{key:0,predicate:i=>i.resourceMeta.type==="Mesh"||V(k.data,i),items:R.toResourceRules},{default:o(({items:i})=>[d("div",K,[(t(!0),n(c,null,f(Object.groupBy(i,r=>r.type),(r,y)=>(t(),n("div",{key:y},[u(A,{"policy-type":y},{default:o(()=>[a(h(y),1)]),_:2},1032,["policy-type"]),e[5]||(e[5]=a()),d("div",Q,[(t(!0),n(c,null,f(r.length>1?r.filter(s=>V(k.data,s)):r,s=>(t(),n("div",{key:s},[s.origins.length>0?(t(),p(M,{key:0,layout:"horizontal"},{title:o(()=>e[2]||(e[2]=[a(`
                                      Origin Policies
                                    `)])),body:o(()=>[u(C,{predicate:l=>typeof l.resourceMeta<"u",items:s.origins,empty:!1},{default:o(({items:l})=>[d("ul",null,[(t(!0),n(c,null,f(l,b=>(t(),n("li",{key:JSON.stringify(b)},[Object.keys(_).length>0?(t(),p(D,{key:0,to:{name:"policy-detail-view",params:{policyPath:_[y][0].path,mesh:b.resourceMeta.mesh,policy:b.resourceMeta.name}}},{default:o(()=>[a(h(b.resourceMeta.name),1)]),_:2},1032,["to"])):g("",!0)]))),128))])]),_:2},1032,["predicate","items"])]),_:2},1024)):g("",!0),e[4]||(e[4]=a()),u(N,{class:"mt-2",code:P($).stringify(s.raw),language:"yaml","show-copy-button":!1},null,8,["code"])]))),128))])]))),128))])]),_:2},1032,["predicate","items"])):(t(),p(C,{key:1,predicate:i=>i.ruleType==="to"&&!["MeshHTTPRoute","MeshTCPRoute"].includes(i.type)&&i.matchers.every(r=>r.key==="kuma.io/service"&&(r.not?r.value!==O:r.value===O)),items:R.rules},{default:o(({items:i})=>[d("div",W,[u(J,{"initially-open":0,"multiple-open":"",class:"stack"},{default:o(()=>[(t(!0),n(c,null,f(Object.groupBy(i,r=>r.type),(r,y)=>(t(),p(S,{key:y},{default:o(()=>[u(U,null,{"accordion-header":o(()=>[u(A,{"policy-type":y},{default:o(()=>[a(h(y)+" ("+h(r.length)+`)
                                  `,1)]),_:2},1032,["policy-type"])]),"accordion-content":o(()=>[d("div",Z,[(t(!0),n(c,null,f(r,s=>(t(),n(c,{key:s},[s.matchers.length>0?(t(),p(M,{key:0,layout:"horizontal"},{title:o(()=>e[6]||(e[6]=[a(`
                                          From
                                        `)])),body:o(()=>[d("p",null,[u(Y,{items:s.matchers},null,8,["items"])])]),_:2},1024)):g("",!0),e[12]||(e[12]=a()),s.origins.length>0?(t(),p(M,{key:1,layout:"horizontal"},{title:o(()=>e[8]||(e[8]=[a(`
                                          Origin Policies
                                        `)])),body:o(()=>[d("ul",null,[(t(!0),n(c,null,f(s.origins,l=>(t(),n("li",{key:`${l.mesh}-${l.name}`},[_[l.type]?(t(),p(D,{key:0,to:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:_[l.type][0].path,policy:l.name}}},{default:o(()=>[a(h(l.name),1)]),_:2},1032,["to"])):(t(),n(c,{key:1},[a(h(l.name),1)],64))]))),128))])]),_:2},1024)):g("",!0),e[13]||(e[13]=a()),d("div",null,[e[10]||(e[10]=d("dt",null,`
                                          Config
                                        `,-1)),e[11]||(e[11]=a()),d("dd",ee,[d("div",null,[u(N,{code:P($).stringify(s.raw),language:"yaml","show-copy-button":!1},null,8,["code"])])])])],64))),128))])]),_:2},1024)]),_:2},1024))),128))]),_:2},1024)])]),_:2},1032,["predicate","items"]))]),_:2},1032,["src"]))),128))]),_:2},1024)])):g("",!0)]))),128))]),_:2},1024)]),_:1},8,["name"])}}}),se=H(te,[["__scopeId","data-v-d7848306"]]);export{se as default};
