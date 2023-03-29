import{d as S,L as ae,o as e,c as t,e as a,F as m,z as Q,g as i,y,a as u,w as n,x as z,k as B,Q as Pe,u as v,_ as ke,A as J,B as W,N as se,O as ne,H as x,i as X,j as R,b as Ee,R as Oe,S as fe,l as ve,r as _,n as te,C as Ue,T as Qe,U as ie,V as Me,J as Ge,W as _e,X as Ie,Y as Re,Z as ze,$ as Le,a0 as Se,a1 as xe,m as Ye}from"./index-e096fb01.js";import{_ as Te}from"./CodeBlock.vue_vue_type_style_index_0_lang-25903ab0.js";import{T as q}from"./TagList-91f2b0bd.js";import{_ as De}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-467e5ff3.js";import{E as be}from"./ErrorBlock-e347ad5a.js";import{_ as Be}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-7adf9901.js";import{t as oe}from"./toYaml-4e00099e.js";import{_ as Ne,E as ee}from"./EnvoyData-e00d5aaf.js";import{_ as Ae}from"./LabelList.vue_vue_type_style_index_0_lang-4dec2f30.js";import{S as He}from"./StatusBadge-79f7109b.js";import{_ as Ke,S as qe}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-bf69e6e6.js";import{T as je}from"./TabsWidget-760376e0.js";import{_ as Fe}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-2a258ae8.js";import{_ as Je}from"./YamlView.vue_vue_type_script_setup_true_lang-8340ad04.js";import"./QueryParameter-70743f73.js";const K=o=>(se("data-v-a67bcff4"),o=o(),ne(),o),We={class:"mesh-gateway-policy-list"},Xe=K(()=>i("h3",null,"Gateway policies",-1)),Ve={key:0,class:"policy-list"},Ze=K(()=>i("h3",{class:"mt-6"},`
      Listeners
    `,-1)),$e=K(()=>i("b",null,"Host",-1)),et=K(()=>i("h4",{class:"mt-2"},`
              Routes
            `,-1)),tt={class:"dataplane-policy-header"},at=K(()=>i("b",null,"Route",-1)),st=K(()=>i("b",null,"Service",-1)),nt={key:0,class:"badge-list"},lt={class:"policy-list mt-1"},it=S({__name:"MeshGatewayDataplanePolicyList",props:{meshGatewayDataplane:{type:Object,required:!0},meshGatewayListenerEntries:{type:Array,required:!0},meshGatewayRoutePolicies:{type:Array,required:!0}},setup(o){const c=o;return(T,E)=>{const f=ae("router-link");return e(),t("div",We,[Xe,a(),o.meshGatewayRoutePolicies.length>0?(e(),t("ul",Ve,[(e(!0),t(m,null,Q(o.meshGatewayRoutePolicies,(p,D)=>(e(),t("li",{key:D},[i("span",null,y(p.type),1),a(`:

        `),u(f,{to:p.route},{default:n(()=>[a(y(p.name),1)]),_:2},1032,["to"])]))),128))])):z("",!0),a(),Ze,a(),i("div",null,[(e(!0),t(m,null,Q(c.meshGatewayListenerEntries,(p,D)=>(e(),t("div",{key:D},[i("div",null,[i("div",null,[$e,a(": "+y(p.hostName)+":"+y(p.port)+" ("+y(p.protocol)+`)
          `,1)]),a(),p.routeEntries.length>0?(e(),t(m,{key:0},[et,a(),u(W,{"initially-open":[],"multiple-open":""},{default:n(()=>[(e(!0),t(m,null,Q(p.routeEntries,(A,b)=>(e(),B(J,{key:b},Pe({"accordion-header":n(()=>[i("div",tt,[i("div",null,[i("div",null,[at,a(": "),u(f,{to:A.route},{default:n(()=>[a(y(A.routeName),1)]),_:2},1032,["to"])]),a(),i("div",null,[st,a(": "+y(A.service),1)])]),a(),A.policies.length>0?(e(),t("div",nt,[(e(!0),t(m,null,Q(A.policies,(s,g)=>(e(),B(v(ke),{key:`${D}-${g}`},{default:n(()=>[a(y(s.type),1)]),_:2},1024))),128))])):z("",!0)])]),_:2},[A.policies.length>0?{name:"accordion-content",fn:n(()=>[i("ul",lt,[(e(!0),t(m,null,Q(A.policies,(s,g)=>(e(),t("li",{key:`${D}-${g}`},[a(y(s.type)+`:

                      `,1),u(f,{to:s.route},{default:n(()=>[a(y(s.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):z("",!0)])]))),128))])])}}});const ot=x(it,[["__scopeId","data-v-a67bcff4"]]),ce="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB4CAMAAAAOusbgAAAAVFBMVEXa2tra2tra2tra2tra2tra2tr////a2toAfd6izPLvzPnRfvDYteSKr86zas0Aar4AhODY6vr3+Prx8v2Kv+9aqOk3muUOj+N5t+211vXhqfW01fXvn55GAAAABnRSTlMC9s/Hbhsvz/I3AAABVklEQVRo3u3b3Y6CMBCG4SJYhnV/KD+K7v3f57bN7AFJTcDUmZB+74lH5EmMA5hmjK+pq1awqm5M6HxqxTudPSzssmxM06rUmDp8DFawIYi1qYRdlisTeCtcMAGnAgwYMGDAgJ8GGPDB4B8frepnl9cZH5d1374E7GmX1WVuA0xzTvixA+5zwpc0/OXrVgU5N/yx6tMHGDBgwIABvxmeiBZhmF3fPMjDFLuOSjDdnBJMvVOAb1G+y8PjlUKdOGyHOcpLJniiDfEVC/FYZYA3unxFx2OVAd7sTjZ073msRGB2Yy7KvcsC2z05Hitx2P6PVTEwf9W/h/5xvTBOB76ByN8ydzRRzofELln1schjVNCrTxyjsl5vtV7ol7L+tAEGDLhMWOAw5ADHPxIHXmpHfAWepgJOBBgwYMCAAT8NMGDAgJOw2hKO2tqR2qKV1mqZ3jKd2vrgH/W3idgykdWgAAAAAElFTkSuQmCC",At="data:image/png;base64,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",re="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEj0lEQVRoBe1aS28TVxT+7ngydhwnPGISTBKHEB6iUtOWHTvWqCtKqQhISC2vBbCpRDf9BUgsgAWbdlGppWqlSl1UXfMLCmXRqgXUxmCclOCWxI4Tv4Zz7s00BntmLh4rTCSfxJ4755458333fHfuTTQCZFOHTo+ijCs2cAi2nWJfaE2InABuw8Lle7e/eCwYvL2CXwF7a2hBtwQm8iKKdwwe+Y0HnhnRgBN2Q8qmJcPwOxm7EXrNe40jzVfDq38j9HUJvOkqdSvQrUDAEeiIhGaPH8bsyfe1oWQuTuPxhePa8V6BplenTl85tQ2l9A7YMUsnHMsTIyjtm9CK1QkKXIHC3nEI2l3RgqhzPzw/sB/g+A5ZYAKlPTsVFMnCH1Xx3f26XP2TUUQgAuXhJKr9fQqQRgVYPpUtA7IANvQq5sciEIHi7jHKb5OE9DQh5SOvoGs6pKNABJYn06tAaDQ1SLB82DoFnnO1TaA8NIhqIo7IQkFLDI58zPx/WvEMTsfaJlAiPbPF789oiWHxPTX6A3f/kPGdmQEBKlCaGJUE+oiANJ9JvEAEeOL23/ldHvVmjUrt9d1WBSrJLaiRfMzCEqzcU8pPcDzmAMunSk8f699FxP7KqngvVK/R19ZKvDy+Qy5cvQ8z8la2xuhzII8+m9foF9+axOz0YRm3/dbP6PvtoWy7fZm1iIV6tAd1i4+W3BLUrR7Y1Jb+1T7eKqg41ccajj94JPPy4DskaoleZM8cRYmeUGyO1hm0Q6DRz5XMnj2KpV1jTcSYyOTnNzjc1Uw1eCwBpQIFhNWqfvhKCZDPZbCQoGK5eVhz82uJKYjBPDp/DFwhBswZnEcmT3YlnzV/jRbBzKVplFNDTeDXEnu3TLNeBpb44x3o20vksh8fQYU2d1GaF+nr3yBCc6SVOaQyl05gxYm/9rWMf1VCra5v9LU1BxoT/N+mCpSHB2HNzmP05neu4J14ltZKKqnIroLnPta8n2ycHHzsHAGqgPXPM4x8+QOBLzXeo6ntSMsiGaYbwDcFajg6QiA6k0M9EQM/NSJFb/CMqe/PDD0QTKrU976V8uMg3j74ifOg8IsNZX9bC1mYmHQJvOlqBJ7EcUPgw8EELFq5vn1WQKHmPaX6IwIXhzdJ3jfmnmPRJ95vgAJJqJfAf0Tgx3pMpGn7cW5oExIE0M0Y/GepzdgT65EfbrPvVZuKW7g6vlV+uO1lYurgWTtmGHIEo7QYxYhSlM6jlJf9UT6nNvtiBFj5+SjUNeRbrNWpLTBmRSiOc6h8bjfOlquya8TyEQDdN1+t4dOZvFsqXsjU3ob/rqVfMv5iGaijbdORO2ihUlshiqdu5RZ4Uqnix3wRBsWcSiawj/8/xAEqGSd8ye4vV8DS4e3EheEBWYmXAl7zJJTrAMvm1LaEpPLV0wLu8V7NxUJJwAVrS3egSdwy4zo7uwTWecCbbtetQNOQrLPDoOd1bp3v2bnbEXZaN+nFiQ1qjJ3WfFymZdN9rQ4tOcJM2CNzf/+ysH33gVuiLlIkpyTh7Q8tZgbGr9sI8RO9qfIBv27zAiEVYZQrGIvuAAAAAElFTkSuQmCC",ue="data:image/png;base64,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",pe="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB5CAYAAADyOOV3AAAFOklEQVR4Ae2dP2tUQRTFp7S385MofgRFiGBhKr9GuqBiI9iktwosCAnRLo0g8Q+ojSaNBomumESTIAqSLiPTTc4k7+bmztudu3sCAztv7p137/nNebtuREMIIXz9eXBluLO/NNzZe8sxCRrsL23tHlxObMP33b3ZzeHO0edv25FjcjRITBPbsPXj12+CnRywOcvENtC9kwk3gU5sQ048vf7775DDsQbIswAc+eNaAQJ2jU8unoBljVxHELBrfHLxBCxr5DqCgF3jk4snYFkj1xEE7BqfXDwByxq5jiBg1/jk4glY1sh1BAG7xicXT8CyRq4jCNg1Prl4ApY1ch1BwK7xycUTsKyR6wgCdo1PLp6AZY1cRxCwa3xy8QQsa+Q6goBd45OLJ2BZI9cRBOwan1w8AcsauY4gYNf45OIJWNbIdQQBu8YnF0/AskauIwjYNT65eAKWNXIdUQD+c2sm5iPemY2mIcnD/bsVMuqTs0yvQ7wQYtXRXb79XtxfpSEB4wH3foCgHwIGQSS+5qeddAOsxzgPOwsLMR9xsBhNQ2qA+3crZNQnZ5le89/o6Jbb3WrxKRovuOuIBR9TAHnSwcfk8T8hYP8MOzsg4E55/C8SsH+GnR0QcKc8/hcJ2D/Dzg4IuFMe/4sE7J9hZwcE3CmP/8WpAzz7cCnm48bdQaw58r3T63H/TB3gcG0+jnIQ8IgVGCXcdK9x/9DBPTuagEesADr43uBFrDlwf217+B5unV+fX4z5mPjfJiGA95vbsebA/bWAMb/6HJ/Z2gJbj0fBasJNe+H+Wj0wv/qcgG2ORiAErFXAGI8AWnfwo5U30TLmHq/GfPA92PiejAdIex4x33oAl9c+xnwQMAFrz2Rb8bUdgg7D/bXdYz7ur53n7k2v6WA6WHsm24qv7RB0FO6v7R7zcX/tnA42OhYFR0AErFXAGI8AEJB1jvtry8V8az2Fg/PvLdNr63ehmK9tuHZ8bQERAO6vrR/zcX/tvACMN6g91zZcOx770QomxeP+2voxX7qftE7A0/YejCeo9lx7omvHYz+SA7TruL+2fszX3h/jCwfn31um15bvQVMuFqxtGN/DrXOsBwWxznF/bb+Yb62nAIwXrDfAgq0N437WubU/zMd6rP3i/to58gx4QbshxtduGPezzrFe6xzrIWBBARSs9twKFPOxPqG9YhnzcX/tHA3bvIOtnwkwXyuYFI+ACoLCBcyX7ietuwMsNTTudQQk8CyWMd/aDwFP25+Dkbj1BOGJLI6scAHzrfX0nY/1Cu0Vy5hvrRd5Nv8ebG2473wEVBAULmC+tV4C5iO6rb9Gaj3RfeejAwXDFsuYb62XDqaD6WCNi9CBhUWFC5ivufdJsXQwHUwHn+SM066hAwXDFsuYf9p9znqdDqaD6zpY+/vc2if6rCf/vHFY77j7HbmDUQDt/LzCjypP248Ub62bgHt+REsApXUCrgzIKqgETLturad3B+PvX61za8N951v7w3xrvb0DthbIfNuHXAJu7BFf+0ATMAHbHhG1TyT30/Ggg+lg3Ymhw9rSiw6mg9s6kXxC6HjQwXSw7sTQYW3pRQfTwW2dSD4hdDzoYDpYd2LosLb0ooPp4LZOJJ8QOh50MB2sOzF0WFt60cF0cFsnkk8IHQ86mA7WnRg6rC296OBpd/Dqu0+Rw68GhYNXXq4f4UXOj//fQ171SGzD8tr60GsDrFs6iOvDcPP+k5mnrzYOKZYklq/1xDSxDWHmwcWr84NLz15v3H7+4csch38NEsvENLH9DwLs1co+Fv2iAAAAAElFTkSuQmCC",de=""+new URL("Retry-8b2ec896.png",import.meta.url).href,ye=""+new URL("Timeout-dcabf0f7.jpg",import.meta.url).href,he="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAABYklEQVRoBe2av0oDQRDGZxbRxhfwDRI0NhKtRAhWPkM6Ex9KTOczWElArBRsAuEeIS+QRpvJfJdcqkWRLWYH5or7s7N797v59j4Odph2m4hw//xywsT3JHQqJMddrIajcq2Jaalcs2bx+cTMAi7Grn9xfSI/388kMsJ19RvznA+Pxs3X+yoh867gkV1NNJjBzr3BcKpT5rH6rOcAmR5SO+dzQQdtYE/4YB2w5hGVPdXmNnnSfCvYUz7kpzVewFor9woc/DeDb/OXX4fcjO728b/67jsWnLhXgHtnw/anqCAJpkPdKxAvYDp/9OHhQtYKhAtZKxAuZK1AuJC1AuFC1gqEC1krEC5krUC4kLUC4ULWCoQLWSsQLmStQLhQKFCYAaxSrgvvYTYc7AnL92YEpQ9WdqxSzkrvYzUe7Lwt8rh6dVMn0WVL6yWaxcdtQtUHCidIG7pY9cddsUfL3sF6LbfZAN5wf/+tIkpkAAAAAElFTkSuQmCC",me="data:image/png;base64,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",ct="data:image/png;base64,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",ge="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEs0lEQVRoBe1azWtUVxQ/781HkslHzQwpDfkQUUpdaHZtaUtTuhACYtC/wI22FHd+bMSlFNSdIhjcddFNKW0pZlfS0BayEdSFqAjRJMbGfBgTZ+JM8p7nd27ezMvkvsy7yUucAS+898479+P8zj3n3nvembGIy8xMttOhwiVy6RuX3HbwqrVYZE2SRUM2Jc5lMqlxaxX8Hdd109UKWofLsqxZVqLHxszXGngoBMzAbsNtdBrWBI+x29Xu8xtNJLDbGzWohbr3CrxrK8W3A4BtW9SYqqdEQg1fKCzT6+wSOY4bubjIFQD41g+ayLZL3hmrS1KSlZmbX4xciZKUiOYGMw/wfz/M0ldXnsgFGjzURV2sfRceF+3KhwPxCYdDQslhml+ImVq54KKlVwv6v7Pd9GFzXIA/f7Ui/T5qidE/Z3bT1MIyfXn5qfRBhb9/ptGmvv11dOLzFCVi0i3ULe560mVEBr/6lN4igW/+Nr5hRU+u8/TlfdlXqychr9QO8tQUTGcd+ul2TmT98EVK31nDtaamX1aWqukYxGpuaqB69nm4zfk/Xkizi0faqPfjFC29ydPCIoPUlH9H83T61gKlUzbdOt6qaaFnRa6AbhFDtOM4FRfxpzdmxNgj32X0aDXcyHchbJXYbTa1jTIa0502cgXUbLuBrqKZxCJrhffEgP2i2Kac2BYFyoWEfmc0pguyqhRwePusaQu4cW9bDW0z2hYLbDYWcmOmDkTRK6DbRsPGQuJC4SdfWm5bLDQ4sURdv07KBbpSLDT8f55c9oc0hxQmxciFCg7RtUdZ+v1ZnqbfOBIz+WMn0HePdhFCtpMjczSe4w6r9NixdprnffLA4CxzAlwlZlF/d530CXszUuDq4yzdfLqkgh+eKMDwLzuhA+ImAEIc5LCfl3YaNFajtNXZ1N+epFN7w8dBGNNIgd+m8gJgoKeFvs4k0H9daeZEDcrAZ61iBY/GcxfX3T8UPkxAn0rFKBb6ZHhW3OZBbzAI3SIGiDCxUCWwunojCzichPHZXzeefHFtOhbSjrgx00gBHDRhCgK6oLA5TH+TNkYKYAFWWzFSgOSory4VjBRQFgjnRjulptGxJ8FWiA9u7ET4tEy3NssFGrytlO9fLNLMynoXNlOAW1daB942iu/iGKdScIFWuaLNK/FnNk/fTr4kPP3FSIG2es7Gs9P99brgH2MN7eWFBl/lqOv+hFygK8VCawYJeIEFYAm/NYwOsh/ncnR9PldMo3hhgHpCqkWjB7uoPRkX4OMFlRfq5ETP2P4Omswv0557Y3IYKoywiAolpDe/+tNQHi1pm7KpznDcdDHdaBZKnNnVwPGMS78s5mlqhUMGBDUiX7mGCFKkwld+R/PVSwDzrQSf3ZPfMaQKRvCCrBEz+Mm/jaHLumJkgXW9NQwvLwS3OTmByJPjoo409bU0bJgX0gy1htX5RI0F5uFUUmYfVjDaRteMGPCCLDQSuQA81tJRbIVYCHVbKZ7bQAGvRK7AlvJCHirN0z/r/urIXcg/+E7QZWt7J0RGK+O9AtHOp/loHKHwfw9qtAC7zefDUI3i5wOOhmr/zx74ywr+9cE5nZ9rwZ2AEViBGdjfAhPs4mowdpbkAAAAAElFTkSuQmCC",rt=""+new URL("VirtualOutbound-3bb05b70.png",import.meta.url).href,ut={class:"policy-type-tag"},pt=["src"],dt=S({__name:"PolicyTypeTag",props:{policyType:{type:String,required:!0}},setup(o){const c=o,T=X(),E={CircuitBreaker:{iconUrl:ce},FaultInjection:{iconUrl:At},HealthCheck:{iconUrl:re},MeshAccessLog:{iconUrl:he},MeshCircuitBreaker:{iconUrl:ce},MeshGateway:{iconUrl:null},MeshGatewayRoute:{iconUrl:null},MeshHealthCheck:{iconUrl:re},MeshProxyPatch:{iconUrl:ue},MeshRateLimit:{iconUrl:pe},MeshRetry:{iconUrl:de},MeshTimeout:{iconUrl:ye},MeshTrace:{iconUrl:ge},MeshTrafficPermission:{iconUrl:me},ProxyTemplate:{iconUrl:ue},RateLimit:{iconUrl:pe},Retry:{iconUrl:de},Timeout:{iconUrl:ye},TrafficLog:{iconUrl:he},TrafficPermission:{iconUrl:me},TrafficRoute:{iconUrl:ct},TrafficTrace:{iconUrl:ge},VirtualOutbound:{iconUrl:rt}},f=R(()=>{const D=T.state.policyTypes.map(A=>{const b=E[A.name]??{iconUrl:null};return[A.name,b]});return Object.fromEntries(D)}),p=R(()=>f.value[c.policyType]);return(D,A)=>(e(),t("span",ut,[v(p).iconUrl!==null?(e(),t("img",{key:0,class:"policy-type-tag-icon",src:v(p).iconUrl,alt:""},null,8,pt)):(e(),B(v(Ee),{key:1,icon:"brain",size:"24"})),a(),Oe(D.$slots,"default",{},()=>[a(y(c.policyType),1)],!0)]))}});const we=x(dt,[["__scopeId","data-v-0052ac03"]]),yt={class:"policy-type-heading"},ht={class:"policy-list"},mt={key:0,class:"origin-list"},gt=S({__name:"PolicyTypeEntryList",props:{id:{type:String,required:!1,default:"entry-list"},policyTypeEntries:{type:Object,required:!0}},setup(o){const c=o,T=[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function E({headerKey:f}){return{class:`cell-${f}`}}return(f,p)=>{const D=ae("router-link");return e(),B(W,{"initially-open":[],"multiple-open":""},{default:n(()=>[(e(!0),t(m,null,Q(c.policyTypeEntries,(A,b)=>(e(),B(J,{key:b},{"accordion-header":n(()=>[i("h3",yt,[u(we,{"policy-type":A.type},{default:n(()=>[a(y(A.type)+" ("+y(A.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":n(()=>[i("div",ht,[u(v(fe),{class:"policy-type-table",fetcher:()=>({data:A.connections,total:A.connections.length}),headers:T,"cell-attrs":E,"disable-pagination":"","is-clickable":""},{sourceTags:n(({rowValue:s})=>[s.length>0?(e(),B(q,{key:0,class:"tag-list",tags:s},null,8,["tags"])):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),destinationTags:n(({rowValue:s})=>[s.length>0?(e(),B(q,{key:0,class:"tag-list",tags:s},null,8,["tags"])):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),name:n(({rowValue:s})=>[s!==null?(e(),t(m,{key:0},[a(y(s),1)],64)):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),origins:n(({rowValue:s})=>[s.length>0?(e(),t("ul",mt,[(e(!0),t(m,null,Q(s,(g,O)=>(e(),t("li",{key:`${b}-${O}`},[u(D,{to:g.route},{default:n(()=>[a(y(g.name),1)]),_:2},1032,["to"])]))),128))])):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),config:n(({rowValue:s,rowKey:g})=>[s!==null?(e(),B(Te,{key:0,id:`${c.id}-${b}-${g}-code-block`,code:s,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const ft=x(gt,[["__scopeId","data-v-e55b8bdf"]]),vt={class:"policy-type-heading"},Tt={class:"policy-list"},Dt={key:1,class:"tag-list-wrapper"},bt={key:0},Bt={key:1},wt={key:0,class:"list"},Ct={key:0,class:"list"},Pt=S({__name:"RuleEntryList",props:{id:{type:String,required:!1,default:"entry-list"},ruleEntries:{type:Object,required:!0}},setup(o){const c=o,T=[{label:"Type",key:"type"},{label:"Addresses",key:"addresses"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function E({headerKey:f}){return{class:`cell-${f}`}}return(f,p)=>{const D=ae("router-link");return e(),B(W,{"initially-open":[],"multiple-open":""},{default:n(()=>[(e(!0),t(m,null,Q(c.ruleEntries,(A,b)=>(e(),B(J,{key:b},{"accordion-header":n(()=>[i("h3",vt,[u(we,{"policy-type":A.type},{default:n(()=>[a(y(A.type)+" ("+y(A.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":n(()=>[i("div",Tt,[u(v(fe),{class:"policy-type-table",fetcher:()=>({data:A.connections,total:A.connections.length}),headers:T,"cell-attrs":E,"disable-pagination":"","is-clickable":""},{type:n(({rowValue:s})=>[s.sourceTags.length===0&&s.destinationTags.length===0?(e(),t(m,{key:0},[a(`
                —
              `)],64)):(e(),t("div",Dt,[s.sourceTags.length>0?(e(),t("div",bt,[a(`
                  From

                  `),u(q,{class:"tag-list",tags:s.sourceTags},null,8,["tags"])])):z("",!0),a(),s.destinationTags.length>0?(e(),t("div",Bt,[a(`
                  To

                  `),u(q,{class:"tag-list",tags:s.destinationTags},null,8,["tags"])])):z("",!0)]))]),addresses:n(({rowValue:s})=>[s.length>0?(e(),t("ul",wt,[(e(!0),t(m,null,Q(s,(g,O)=>(e(),t("li",{key:`${b}-${O}`},y(g),1))),128))])):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),origins:n(({rowValue:s})=>[s.length>0?(e(),t("ul",Ct,[(e(!0),t(m,null,Q(s,(g,O)=>(e(),t("li",{key:`${b}-${O}`},[u(D,{to:g.route},{default:n(()=>[a(y(g.name),1)]),_:2},1032,["to"])]))),128))])):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),config:n(({rowValue:s,rowKey:g})=>[s!==null?(e(),B(Te,{key:0,id:`${c.id}-${b}-${g}-code-block`,code:s,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),t(m,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const kt=x(Pt,[["__scopeId","data-v-7558548e"]]),Ce=o=>(se("data-v-ed201f38"),o=o(),ne(),o),Et=Ce(()=>i("h2",{class:"visually-hidden"},`
    Policies
  `,-1)),Ot={key:0,class:"mt-2"},Ut=Ce(()=>i("h2",null,"Rules",-1)),Qt=S({__name:"SidecarDataplanePolicyList",props:{dppName:{type:String,required:!0},policyTypeEntries:{type:Object,required:!0},ruleEntries:{type:Array,required:!0}},setup(o){const c=o;return(T,E)=>(e(),t(m,null,[Et,a(),u(ft,{id:"policies","policy-type-entries":c.policyTypeEntries},null,8,["policy-type-entries"]),a(),o.ruleEntries.length>0?(e(),t("div",Ot,[Ut,a(),u(kt,{id:"rules","rule-entries":c.ruleEntries},null,8,["rule-entries"])])):z("",!0)],64))}});const Mt=x(Qt,[["__scopeId","data-v-ed201f38"]]),Gt={key:2,class:"policies-list"},_t={key:3,class:"policies-list"},It=S({__name:"DataplanePolicies",props:{dataPlane:{type:Object,required:!0}},setup(o){const c=o,T=ve(),E=X(),f=_(null),p=_([]),D=_([]),A=_([]),b=_([]),s=_(!0),g=_(null);te(()=>c.dataPlane.name,function(){O()}),O();async function O(){var l,d;g.value=null,s.value=!0,p.value=[],D.value=[],A.value=[],b.value=[];try{if(((d=(l=c.dataPlane.networking.gateway)==null?void 0:l.type)==null?void 0:d.toUpperCase())==="BUILTIN")f.value=await T.getMeshGatewayDataplane({mesh:c.dataPlane.mesh,name:c.dataPlane.name}),A.value=j(f.value),b.value=F(f.value.policies);else{const{items:r}=await T.getSidecarDataplanePolicies({mesh:c.dataPlane.mesh,name:c.dataPlane.name});p.value=Z(r??[]);const{items:w}=await T.getDataplaneRules({mesh:c.dataPlane.mesh,name:c.dataPlane.name});D.value=P(w??[])}}catch(h){h instanceof Error?g.value=h:console.error(h)}finally{s.value=!1}}function j(l){const d=[];for(const h of l.listeners)for(const r of h.hosts)for(const w of r.routes){const U=[];for(const G of w.destinations){const I=F(G.policies),C={routeName:w.route,route:{name:"policy-detail-view",params:{mesh:l.gateway.mesh,policyPath:"meshgatewayroutes",policy:w.route}},service:G.tags["kuma.io/service"],policies:I};U.push(C)}d.push({protocol:h.protocol,port:h.port,hostName:r.hostName,routeEntries:U})}return d}function F(l){if(l===void 0)return[];const d=[];for(const h of Object.values(l)){const r=E.state.policyTypesByName[h.type];d.push({type:h.type,name:h.name,route:{name:"policy-detail-view",params:{mesh:h.mesh,policyPath:r.path,policy:h.name}}})}return d}function Z(l){const d=new Map;for(const r of l){const{type:w,service:U}=r,G=typeof U=="string"&&U!==""?[{label:"kuma.io/service",value:U}]:[],I=w==="inbound"||w==="outbound"?r.name:null;for(const[C,Y]of Object.entries(r.matchedPolicies)){d.has(C)||d.set(C,{type:C,connections:[]});const N=d.get(C),H=E.state.policyTypesByName[C];for(const le of Y){const L=M(le,H,r,G,I);N.connections.push(...L)}}}const h=Array.from(d.values());return h.sort((r,w)=>r.type.localeCompare(w.type)),h}function M(l,d,h,r,w){const U=l.conf&&Object.keys(l.conf).length>0?oe(l.conf):null,I=[{name:l.name,route:{name:"policy-detail-view",params:{mesh:l.mesh,policyPath:d.path,policy:l.name}}}],C=[];if(h.type==="inbound"&&Array.isArray(l.sources))for(const{match:Y}of l.sources){const H={sourceTags:[{label:"kuma.io/service",value:Y["kuma.io/service"]}],destinationTags:r,name:w,config:U,origins:I};C.push(H)}else{const N={sourceTags:[],destinationTags:r,name:w,config:U,origins:I};C.push(N)}return C}function P(l){const d=new Map;for(const r of l){d.has(r.policyType)||d.set(r.policyType,{type:r.policyType,connections:[]});const w=d.get(r.policyType),U=E.state.policyTypesByName[r.policyType],G=k(r,U);w.connections.push(...G)}const h=Array.from(d.values());return h.sort((r,w)=>r.type.localeCompare(w.type)),h}function k(l,d){const{type:h,service:r,subset:w,conf:U}=l,G=w?Object.entries(w):[];let I,C;h==="ClientSubset"?G.length>0?I=G.map(([L,$])=>({label:L,value:$})):I=[{label:"kuma.io/service",value:"*"}]:I=[],h==="DestinationSubset"?G.length>0?C=G.map(([L,$])=>({label:L,value:$})):typeof r=="string"&&r!==""?C=[{label:"kuma.io/service",value:r}]:C=[{label:"kuma.io/service",value:"*"}]:h==="ClientSubset"&&typeof r=="string"&&r!==""?C=[{label:"kuma.io/service",value:r}]:C=[];const Y=l.addresses??[],N=U&&Object.keys(U).length>0?oe(U):null,H=[];for(const L of l.origins)H.push({name:L.name,route:{name:"policy-detail-view",params:{mesh:L.mesh,policyPath:d.path,policy:L.name}}});return[{type:{sourceTags:I,destinationTags:C},addresses:Y,config:N,origins:H}]}return(l,d)=>s.value?(e(),B(Be,{key:0})):g.value!==null?(e(),B(be,{key:1,error:g.value},null,8,["error"])):p.value.length>0?(e(),t("div",Gt,[u(Mt,{"dpp-name":o.dataPlane.name,"policy-type-entries":p.value,"rule-entries":D.value},null,8,["dpp-name","policy-type-entries","rule-entries"])])):A.value.length>0&&f.value!==null?(e(),t("div",_t,[u(ot,{"mesh-gateway-dataplane":f.value,"mesh-gateway-listener-entries":A.value,"mesh-gateway-route-policies":b.value},null,8,["mesh-gateway-dataplane","mesh-gateway-listener-entries","mesh-gateway-route-policies"])])):(e(),B(De,{key:4}))}});const Rt=x(It,[["__scopeId","data-v-f9acf0cf"]]),V=o=>(se("data-v-ff336bb4"),o=o(),ne(),o),zt={class:"entity-heading"},Lt={key:0},St=V(()=>i("h4",null,"Status",-1)),xt={key:1},Yt=V(()=>i("h4",null,"Reason",-1)),Nt={key:0},Ht=V(()=>i("h4",null,"Tags",-1)),Kt={key:1},qt=V(()=>i("h4",null,"Versions",-1)),jt={key:0},Ft=["href"],Jt=S({__name:"DataPlaneDetails",props:{dataPlane:{type:Object,required:!0},dataPlaneOverview:{type:Object,required:!0}},setup(o){const c=o,T=Ue(),E=X(),f=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"DPP Insights"},{hash:"#dpp-policies",title:"Policies"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"},{hash:"#mtls",title:"Certificate Insights"},{hash:"#warnings",title:"Warnings"}],p=_([]),D=R(()=>{const{type:M,name:P,mesh:k}=c.dataPlane;return{type:M,name:P,mesh:k}}),A=R(()=>Qe(c.dataPlane,c.dataPlaneOverview.dataplaneInsight)),b=R(()=>ie(c.dataPlane)),s=R(()=>Me(c.dataPlaneOverview.dataplaneInsight)),g=R(()=>Ge(c.dataPlane)),O=R(()=>_e(c.dataPlaneOverview)),j=R(()=>{var P;const M=Array.from(((P=c.dataPlaneOverview.dataplaneInsight)==null?void 0:P.subscriptions)??[]);return M.reverse(),M}),F=R(()=>p.value.length===0?f.filter(M=>M.hash!=="#warnings"):f);function Z(){var l;const M=((l=c.dataPlaneOverview.dataplaneInsight)==null?void 0:l.subscriptions)??[];if(M.length===0||!("version"in M[0]))return;const P=M[0].version;if(P&&P.kumaDp&&P.envoy){const d=Re(P);d.kind!==ze&&d.kind!==Le&&p.value.push(d)}E.getters["config/getMulticlusterStatus"]&&P&&ie(c.dataPlane).find(r=>r.label===Se)&&typeof P.kumaDp.kumaCpCompatible=="boolean"&&!P.kumaDp.kumaCpCompatible&&p.value.push({kind:xe,payload:{kumaDp:P.kumaDp.version}})}return Z(),(M,P)=>(e(),B(je,{tabs:v(F)},{tabHeader:n(()=>[i("h1",zt,`
        DPP: `+y(o.dataPlane.name),1)]),overview:n(()=>[u(Ae,null,{default:n(()=>[i("div",null,[i("ul",null,[(e(!0),t(m,null,Q(v(D),(k,l)=>(e(),t("li",{key:l},[i("h4",null,y(l),1),a(),i("div",null,y(k),1)]))),128)),a(),v(A).status?(e(),t("li",Lt,[St,a(),i("div",null,[u(He,{status:v(A).status},null,8,["status"])])])):z("",!0),a(),v(A).reason.length>0?(e(),t("li",xt,[Yt,a(),i("div",null,[(e(!0),t(m,null,Q(v(A).reason,(k,l)=>(e(),t("div",{key:l,class:"reason"},y(k),1))),128))])])):z("",!0)])]),a(),i("div",null,[i("ul",null,[v(b).length>0?(e(),t("li",Nt,[Ht,a(),u(q,{tags:v(b)},null,8,["tags"])])):z("",!0),a(),v(s)?(e(),t("li",Kt,[qt,a(),i("p",null,[(e(!0),t(m,null,Q(v(s),(k,l)=>(e(),t("span",{key:l,class:"tag-cols"},[i("span",null,y(l)+`:
                  `,1),a(),i("span",null,y(k),1)]))),128))])])):z("",!0)])])]),_:1}),a(),u(Je,{id:"code-block-data-plane",class:"mt-4",content:v(g),"is-searchable":""},null,8,["content"])]),insights:n(()=>[u(Ne,{"is-empty":v(j).length===0},{default:n(()=>[u(W,{"initially-open":0},{default:n(()=>[(e(!0),t(m,null,Q(v(j),(k,l)=>(e(),B(J,{key:l},{"accordion-header":n(()=>[u(Ke,{details:k},null,8,["details"])]),"accordion-content":n(()=>[u(qe,{details:k,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1},8,["is-empty"])]),"dpp-policies":n(()=>[u(Rt,{"data-plane":o.dataPlane},null,8,["data-plane"])]),"xds-configuration":n(()=>[u(ee,{"data-path":"xds",mesh:o.dataPlane.mesh,"dpp-name":o.dataPlane.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),"envoy-stats":n(()=>[u(ee,{"data-path":"stats",mesh:o.dataPlane.mesh,"dpp-name":o.dataPlane.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),"envoy-clusters":n(()=>[u(ee,{"data-path":"clusters",mesh:o.dataPlane.mesh,"dpp-name":o.dataPlane.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),mtls:n(()=>[u(Ae,null,{default:n(()=>[v(O)!==null?(e(),t("ul",jt,[(e(!0),t(m,null,Q(v(O),(k,l)=>(e(),t("li",{key:l},[i("h4",null,y(k.label),1),a(),i("p",null,y(k.value),1)]))),128))])):(e(),B(v(Ie),{key:1,appearance:"danger"},{alertMessage:n(()=>[a(`
            This data plane proxy does not yet have mTLS configured —
            `),i("a",{href:`${v(T)("KUMA_DOCS_URL")}/policies/mutual-tls/?${v(T)("KUMA_UTM_QUERY_PARAMS")}`,class:"external-link",target:"_blank"},`
              Learn About Certificates in `+y(v(T)("KUMA_PRODUCT_NAME")),9,Ft)]),_:1}))]),_:1})]),warnings:n(()=>[u(Fe,{warnings:p.value},null,8,["warnings"])]),_:1},8,["tabs"]))}});const Wt=x(Jt,[["__scopeId","data-v-ff336bb4"]]),Xt={class:"component-frame"},pa=S({__name:"DataPlaneDetailView",setup(o){const c=ve(),T=Ye(),E=X(),f=_(null),p=_(null),D=_(!0),A=_(null);async function b(){A.value=null,D.value=!0;const s=T.params.mesh,g=T.params.dataPlane;try{f.value=await c.getDataplaneFromMesh({mesh:s,name:g}),p.value=await c.getDataplaneOverviewFromMesh({mesh:s,name:g})}catch(O){f.value=null,O instanceof Error?A.value=O:console.error(O)}finally{D.value=!1}}return te(()=>T.params.mesh,function(){T.name==="data-plane-detail-view"&&b()}),te(()=>T.params.dataPlane,function(){T.name==="data-plane-detail-view"&&b()}),b(),E.dispatch("updatePageTitle",T.params.dataPlane),(s,g)=>(e(),t("div",Xt,[D.value?(e(),B(Be,{key:0})):A.value!==null?(e(),B(be,{key:1,error:A.value},null,8,["error"])):f.value===null||p.value===null?(e(),B(De,{key:2})):(e(),B(Wt,{key:3,"data-plane":f.value,"data-plane-overview":p.value},null,8,["data-plane","data-plane-overview"]))]))}});export{pa as default};
